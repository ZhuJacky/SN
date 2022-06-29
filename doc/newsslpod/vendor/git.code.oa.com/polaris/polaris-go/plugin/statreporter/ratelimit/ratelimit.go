/**
 * Tencent is pleased to support the open source community by making CL5 available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissionsr and limitations under the License.
 */

package ratelimit

import (
	"context"
	"fmt"
	sysconfig "git.code.oa.com/polaris/polaris-go/pkg/config"
	"git.code.oa.com/polaris/polaris-go/pkg/flow/data"
	"git.code.oa.com/polaris/polaris-go/pkg/flow/quota"
	"git.code.oa.com/polaris/polaris-go/pkg/log"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	"git.code.oa.com/polaris/polaris-go/pkg/network"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/common"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/localregistry"
	"git.code.oa.com/polaris/polaris-go/plugin/statreporter/pb/util"
	monitorpb "git.code.oa.com/polaris/polaris-go/plugin/statreporter/pb/v1"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	trafficShapingListName = "trafficShapingAlgorithm"
)

//限流日志上报插件
type Reporter struct {
	*plugin.PluginBase
	*common.RunContext
	config            *Config
	connectionManager network.ConnectionManager
	connection        *network.Connection
	rateLimitClient   monitorpb.GrpcAPI_CollectRateLimitRecordClient
	clientCancel      context.CancelFunc
	uploadToMonitor   bool
	registry          localregistry.LocalRegistry
	sdkToken          model.SDKToken
	deletedWindow     *sync.Map
	noRuleServices    *sync.Map
}

//记录没有命中规则的服务限流记录
type noRuleRequests struct {
	requests int64
	deleted  uint32
}

//插件类型
func (s *Reporter) Type() common.Type {
	return common.TypeStatReporter
}

//插件名称
func (s *Reporter) Name() string {
	return "rateLimitRecord"
}

// enable
func (g *Reporter) IsEnable(cfg sysconfig.Configuration) bool {
	if cfg.GetGlobal().GetSystem().GetMode() == model.ModeWithAgent {
		return false
	} else {
		for _, name := range cfg.GetGlobal().GetStatReporter().GetChain() {
			if name == g.Name() {
				return true
			}
		}
	}
	return false
}

// destroy 解决匿名组合中该函数二义性问题
func (s *Reporter) Destroy() error {
	err := s.PluginBase.Destroy()
	if err != nil {
		return err
	}
	err = s.RunContext.Destroy()
	if err != nil {
		return err
	}
	return nil
}

//初始化插件
func (s *Reporter) Init(ctx *plugin.InitContext) error {
	s.RunContext = common.NewRunContext()
	s.connectionManager = ctx.ConnManager
	s.PluginBase = plugin.NewPluginBase(ctx)
	s.config = &Config{}
	cfgValue := ctx.Config.GetGlobal().GetStatReporter().GetPluginConfig(s.Name())
	if nil == cfgValue {
		return model.NewSDKError(model.ErrCodeAPIInvalidConfig, nil,
			"config of statReporter rateLimitRecord must be provided")
	}
	s.config = cfgValue.(*Config)
	s.config.SetDefault()
	err := s.config.Verify()
	if err != nil {
		return err
	}
	ctx.Plugins.RegisterEventSubscriber(common.OnRateLimitWindowCreated,
		common.PluginEventHandler{Callback: s.createRateLimitWindowStat})
	ctx.Plugins.RegisterEventSubscriber(common.OnServiceDeleted,
		common.PluginEventHandler{Callback: s.deleteNoRuleService})
	s.registry, _ = data.GetRegistry(ctx.Config, ctx.Plugins)
	t, _ := ctx.ValueCtx.GetValue(model.ContextKeyToken)
	s.sdkToken = t.(model.SDKToken)
	s.deletedWindow = &sync.Map{}
	s.noRuleServices = &sync.Map{}
	return nil
}

// 启动上报协程
func (s *Reporter) Start() error {
	go s.uploadRateLimitRecord()
	return nil
}

//上报限流发生事件
func (s *Reporter) ReportStat(t model.MetricType, info model.InstanceGauge) error {
	if t != model.RateLimitStat {
		return nil
	}
	gauge := info.(*quota.RateLimitGauge)

	window := gauge.Window

	if gauge.Type == quota.WindowDeleted {
		s.deletedWindow.Store(window.Rule.GetRevision().GetValue(), window)
		return nil
	}
	if window == nil {
		svcKey := model.ServiceKey{
			Namespace: gauge.Namespace,
			Service:   gauge.Service,
		}
		if rec, ok := s.noRuleServices.Load(svcKey); ok {
			atomic.AddInt64(&rec.(*noRuleRequests).requests, 1)
		} else {
			newRec, loaded := s.noRuleServices.LoadOrStore(svcKey, &noRuleRequests{
				requests: 1,
				deleted:  0,
			})
			if loaded {
				atomic.AddInt64(&newRec.(*noRuleRequests).requests, 1)
			}
		}
		return nil
	}

	sd := window.PluginData[s.ID()].(*statData)

	switch gauge.Type {
	case quota.TrafficShapingLimited:
		atomic.AddInt64(&sd.trafficShapingLimited.limitedNum, 1)
		atomic.AddInt64(&sd.rejectNum, 1)
	case quota.QuotaLimited:
		atomic.AddInt64(&sd.amountLimited[gauge.AmountIndex].limitedNum, 1)
		atomic.AddInt64(&sd.rejectNum, 1)
	case quota.QuotaRequested:
		atomic.AddInt64(&sd.totalNum, 1)
	case quota.QuotaGranted:
		atomic.AddInt64(&sd.passNum, 1)
	}

	return nil
}

//定时上报限流记录
func (s *Reporter) uploadRateLimitRecord() {
	t := time.NewTicker(*s.config.ReportInterval)
	defer t.Stop()
	for {
		select {
		case <-s.Done():
			if nil != s.clientCancel {
				s.clientCancel()
			}
			log.GetBaseLogger().Infof("uploadRateLimitRecord of rateLimitRecord stat_monitor has been terminated")
			return
		case <-t.C:
			s.uploadToMonitor = true
			timeStart := time.Now()
			deadline := timeStart.Add(*s.config.ReportInterval)
			err := s.connectToMonitor(deadline)
			if nil != err {
				log.GetStatReportLogger().Errorf("fail to connect to monitor to report ratelimit Record, error %v", err)
				s.uploadToMonitor = false
			}
			s.iterateRateLimitRecord()
			if s.uploadToMonitor {
				s.closeConnection()
			}
		}
	}
}

//连接monitor
func (s *Reporter) connectToMonitor(deadline time.Time) error {
	var err error
	s.connection, err = s.connectionManager.GetConnection("ReportRateLimit", sysconfig.MonitorCluster)
	if nil != err {
		log.GetStatReportLogger().Errorf("fail to connect to monitor, err: %s", err.Error())
		return err
	}
	client := monitorpb.NewGrpcAPIClient(network.ToGRPCConn(s.connection.Conn))
	var clientCtx context.Context
	clientCtx, s.clientCancel = context.WithDeadline(context.Background(), deadline)
	s.rateLimitClient, err = client.CollectRateLimitRecord(clientCtx)
	if nil != err {
		log.GetStatReportLogger().Errorf("fail to create stream to report ratelimit record, err: %s", err.Error())
		s.closeConnection()
		return err
	}
	return nil
}

//关闭连接
func (s *Reporter) closeConnection() {
	s.clientCancel()
	s.clientCancel = nil
	if s.rateLimitClient != nil {
		s.rateLimitClient.CloseSend()
		s.rateLimitClient = nil
	}
	s.connection.Release("ReportRateLimit")
}

//遍历所有的限流窗口，并上报限流记录
func (s *Reporter) iterateRateLimitRecord() {
	//将那些已经被删除的限流窗口的记录上传
	s.iterateWindowMap(s.deletedWindow, true)

	services := s.registry.GetServices()
	for svcValue := range services {
		svc := svcValue.(model.ServiceKey)
		svcInstances := s.registry.GetInstances(&svc, false, true)
		if !svcInstances.IsInitialized() {
			continue
		}

		rateLimitWindowSet := quota.GetRateLimitWindowSet(svcInstances)
		if rateLimitWindowSet == nil {
			continue
		}
		//将这个服务现有的限流窗口的记录上传
		s.iterateWindowMap(rateLimitWindowSet.WindowByRule, false)
	}

	//将那些没有匹配到限流规则的请求进行上报
	s.iterateNoRuleService()
}

//上报没有匹配到限流规则的请求
func (s *Reporter) iterateNoRuleService() {
	record := s.createEmptyRecord()
	happenTime := time.Now()
	s.noRuleServices.Range(func(k, v interface{}) bool {
		svcKey := k.(model.ServiceKey)
		record.Service = svcKey.Service
		record.Id = uuid.New().String()
		record.Namespace = svcKey.Namespace
		stats := v.(*noRuleRequests)

		totalNum := GetAtomicInt64(&stats.requests)
		record.RequestsCount = &monitorpb.LimitRequestsCount{
			Time: &timestamp.Timestamp{
				Seconds: happenTime.Unix(),
				Nanos:   int32(happenTime.Nanosecond()),
			},
			TotalRequests: uint32(totalNum),
			PassRequests:  uint32(totalNum),
		}
		s.sendRateLimitRecord(record)
		if atomic.LoadUint32(&stats.deleted) == 1 {
			s.noRuleServices.Delete(k)
		}
		return true
	})
}

//创建一个空记录
func (s *Reporter) createEmptyRecord() *monitorpb.RateLimitRecord {
	record := &monitorpb.RateLimitRecord{
		Id:          "",
		SdkToken:    nil,
		RuleId:      "",
		Subset:      "",
		RateLimiter: "",
		LimitStats:  nil,
	}
	record.SdkToken = util.GetPBSDkToken(s.sdkToken)
	if record.SdkToken.Ip == "" {
		record.SdkToken.Ip = s.connectionManager.GetClientInfo().GetIPString()
	}
	return record
}

func (s *Reporter) iterateWindowMap(windowMap *sync.Map, deleteWindow bool) {
	record := s.createEmptyRecord()
	happenTime := time.Now()
	windowMap.Range(func(key, value interface{}) bool {
		window := value.(*quota.RateLimitWindow)
		sd := window.PluginData[s.ID()].(*statData)
		record.LimitStats = nil
		record.RequestsCount = nil

		totalNum := GetAtomicInt64(&sd.totalNum)
		passNum := GetAtomicInt64(&sd.passNum)
		rejectNum := GetAtomicInt64(&sd.rejectNum)
		if totalNum > 0 {
			record.RequestsCount = &monitorpb.LimitRequestsCount{
				Time: &timestamp.Timestamp{
					Seconds: happenTime.Unix(),
					Nanos:   int32(happenTime.Nanosecond()),
				},
				TotalRequests:  uint32(totalNum),
				PassRequests:   uint32(passNum),
				RejectRequests: uint32(rejectNum),
			}
		}

		trafficLimitNum := GetAtomicInt64(&sd.trafficShapingLimited.limitedNum)
		//atomic.LoadInt64(&sd.trafficShapingLimited.limitedNum)
		if trafficLimitNum > 0 {
			record.LimitStats = append(record.LimitStats, &monitorpb.LimitStat{
				Time: &timestamp.Timestamp{
					Seconds: happenTime.Unix(),
					Nanos:   int32(happenTime.Nanosecond()),
				},
				PeriodTimes: uint32(trafficLimitNum),
				Reason:      sd.trafficShapingLimited.reason,
			})
			//atomic.AddInt64(&sd.trafficShapingLimited.limitedNum, -trafficLimitNum)
		}

		for _, amountStat := range sd.amountLimited {
			amountLimitNum := GetAtomicInt64(&amountStat.limitedNum)
			if amountLimitNum > 0 {
				record.LimitStats = append(record.LimitStats, &monitorpb.LimitStat{
					Time: &timestamp.Timestamp{
						Seconds: happenTime.Unix(),
						Nanos:   int32(happenTime.Nanosecond()),
					},
					PeriodTimes: uint32(amountLimitNum),
					Reason:      amountStat.reason,
				})
				//atomic.AddInt64(&amountStat.limitedNum, -amountLimitNum)
			}
		}
		if len(record.LimitStats) == 0 {
			return true
		}

		record.Id = uuid.New().String()
		record.RuleId = window.Rule.GetId().GetValue()
		record.Subset = window.Rule.GetCluster().GetValue()
		record.RateLimiter = window.Rule.GetAction().GetValue()
		record.Namespace = window.SvcKey.Namespace
		record.Service = window.SvcKey.Service
		record.Labels = sd.ruleMatchLabels

		s.sendRateLimitRecord(record)

		if deleteWindow {
			windowMap.Delete(key)
		}
		return true
	})
}

func (s *Reporter) sendRateLimitRecord(record *monitorpb.RateLimitRecord) {
	//打印到statLog
	log.GetStatLogger().Infof("sdk ratelimit record:%v", record)
	if !s.uploadToMonitor {
		log.GetStatReportLogger().Warnf("Skip to report ratelimit record to monitor for connection problem,"+
			" id: %s", record.Id)
		return
	}

	//上传到monitor
	err := s.rateLimitClient.Send(record)
	if nil != err {
		log.GetStatReportLogger().Errorf("fail to report ratelimit record, id: %s, err %s, monitor server is %s",
			record.Id, err.Error(), s.connection.ConnID)
	}
	resp, err := s.rateLimitClient.Recv()
	if nil != err || resp.Id.GetValue() != record.Id || resp.Code.GetValue() != monitorpb.ReceiveSuccess {
		log.GetStatReportLogger().Errorf("fail to report ratelimit record, resp is %v, err is %v, monitor server is %s",
			resp, err, s.connection.ConnID)
	} else {
		log.GetStatReportLogger().Infof("Success to report ratelimit record, resp is %v, monitor server is %s",
			resp, s.connection.ConnID)
	}
}

//从一个timeList获取限流事件列表
//func getEventsFromList(record *monitorpb.RateLimitRecord, list *timeList, nameAsReason bool) {
//	recordList := list.getTimes()
//	if recordList != nil {
//		for recordList != nil {
//			if nameAsReason {
//				recordList.event.Reason = list.name
//			}
//			record.Events = append(record.Events, recordList.event)
//			recordList = recordList.nextTime
//		}
//	}
//}

//在限流窗口里面创建统计数据
func (s *Reporter) createRateLimitWindowStat(event *common.PluginEvent) error {
	window := event.EventObject.(*quota.RateLimitWindow)
	sd := &statData{}
	sd.trafficShapingLimited = &limitedStat{
		reason: "rateLimiter: " + window.Rule.GetAction().GetValue(),
	}

	sd.amountLimited = make([]*limitedStat, len(window.Rule.GetAmounts()))
	for idx, duration := range window.Rule.GetAmounts() {
		sd.amountLimited[idx] = &limitedStat{
			reason:        fmt.Sprintf("amount:%d/%ds", duration.MaxAmount.GetValue(), duration.ValidDuration.Seconds),
			validDuration: duration.ValidDuration.GetSeconds(),
		}
	}
	sort.Sort(limitedSlice(sd.amountLimited))
	sd.ruleMatchLabels = MarshalRateLimitRuleLabels(window.Rule.Labels)
	//sd.ruleType = ruleTypesMap[window.Rule.GetType()][window.Rule.GetResource()]
	window.PluginData[s.ID()] = sd
	return nil
}

//删除淘汰掉的无规则服务
func (s *Reporter) deleteNoRuleService(event *common.PluginEvent) error {
	svcEventObj := event.EventObject.(*common.ServiceEventObject)
	if svcEventObj.SvcEventKey.Type == model.EventInstances {
		rec, ok := s.noRuleServices.Load(model.ServiceKey{
			Namespace: svcEventObj.SvcEventKey.Namespace,
			Service:   svcEventObj.SvcEventKey.Service,
		})
		if ok {
			atomic.StoreUint32(&rec.(*noRuleRequests).deleted, 1)
		}
	}
	return nil
}

//注册插件和配置
func init() {
	plugin.RegisterConfigurablePlugin(&Reporter{}, &Config{})
}
