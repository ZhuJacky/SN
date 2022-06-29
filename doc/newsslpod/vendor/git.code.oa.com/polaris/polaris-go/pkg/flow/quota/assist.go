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
 * specific language governing permissions and limitations under the License.
 */

package quota

import (
	"git.code.oa.com/polaris/polaris-go/pkg/clock"
	"git.code.oa.com/polaris/polaris-go/pkg/config"
	"git.code.oa.com/polaris/polaris-go/pkg/flow/data"
	"git.code.oa.com/polaris/polaris-go/pkg/log"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	"git.code.oa.com/polaris/polaris-go/pkg/model/local"
	"git.code.oa.com/polaris/polaris-go/pkg/model/pb"
	namingpb "git.code.oa.com/polaris/polaris-go/pkg/model/pb/v1"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/common"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/localregistry"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/serverconnector"
	"github.com/modern-go/reflect2"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Disabled      = "rateLimit disabled"
	RuleNotExists = "quota rule not exists"
)

var MetricReportGlobalMsgId int64 = 0

//限额流程的辅助类
type FlowQuotaAssistant struct {
	//是否启用限流，如果不启用，默认都会放通
	enable bool
	//流程执行引擎
	engine model.Engine
	//服务实例注册中心
	registry localregistry.InstancesRegistry
	//插件工厂
	supplier plugin.Supplier
	//限流server连接器
	rLimitConnector      serverconnector.RateLimitConnector
	asyncRLimitConnector serverconnector.AsyncRateLimitConnector
	//任务列表
	taskValues model.TaskValues

	destroyed uint32

	configMode       model.ConfigMode
	rateLimitCluster config.ServerClusterConfig

	windowCount   int32
	maxWindowSize int32

	windowCountLogCtrl uint64
}

func (f *FlowQuotaAssistant) Destroy() {
	atomic.StoreUint32(&f.destroyed, 1)
	f.asyncRLimitConnector.Destroy()
}

func (f *FlowQuotaAssistant) IsDestroyed() bool {
	return atomic.LoadUint32(&f.destroyed) > 0
}

func (f *FlowQuotaAssistant) AddWindowCount() {
	atomic.AddInt32(&f.windowCount, 1)
}

func (f *FlowQuotaAssistant) DelWindowCount() {
	atomic.AddInt32(&f.windowCount, -1)
}

func (f *FlowQuotaAssistant) GetWindowCount() int32 {
	return atomic.LoadInt32(&f.windowCount)
}

//初始化限额辅助
func (f *FlowQuotaAssistant) Init(engine model.Engine, cfg config.Configuration, supplier plugin.Supplier) error {
	f.engine = engine
	f.supplier = supplier
	connector, err := data.GetServerConnector(cfg, supplier)
	if nil != err {
		return err
	}
	f.rLimitConnector = connector.GetRateLimitConnector()
	f.asyncRLimitConnector = connector.GetAsyncRateLimitConnector()
	registry, err := data.GetRegistry(cfg, supplier)
	if nil != err {
		return err
	}
	f.registry = registry
	f.enable = cfg.GetProvider().GetRateLimit().IsEnable()
	if !f.enable {
		return nil
	}
	callback, err := NewRemoteQuotaCallback(cfg, supplier, engine)
	if nil != err {
		return err
	}
	period := config.MinRateLimitReportInterval
	_, taskValues := engine.ScheduleTask(&model.PeriodicTask{
		Name:       "quota-metric",
		CallBack:   callback,
		Period:     period,
		DelayStart: true,
	})
	f.taskValues = taskValues
	supplier.RegisterEventSubscriber(common.OnServiceLocalValueCreated,
		common.PluginEventHandler{Callback: f.createQuotaLocalValue})
	supplier.RegisterEventSubscriber(common.OnServiceUpdated,
		common.PluginEventHandler{Callback: f.OnServiceUpdated})
	supplier.RegisterEventSubscriber(common.OnServiceDeleted,
		common.PluginEventHandler{Callback: f.OnServiceDeleted})
	f.destroyed = 0
	f.windowCount = 0
	f.maxWindowSize = int32(cfg.GetProvider().GetRateLimit().GetMaxWindowSize())

	f.configMode = model.ConfigMode(cfg.GetProvider().GetRateLimit().GetMode())
	f.windowCountLogCtrl = 0

	metricCluster := cfg.GetProvider().GetRateLimit().GetRateLimitCluster()
	if metricCluster != nil && !reflect2.IsNil(metricCluster) && metricCluster.GetNamespace() != "" &&
		metricCluster.GetService() != "" {
		f.rateLimitCluster = cfg.GetProvider().GetRateLimit().GetRateLimitCluster()
	} else {
		f.rateLimitCluster = nil
	}
	return nil
}

//创建配额本地缓存
func (f *FlowQuotaAssistant) createQuotaLocalValue(event *common.PluginEvent) error {
	lv := event.EventObject.(local.ServiceLocalValue)
	windowSet := &RateLimitWindowSet{WindowByRule: &sync.Map{}, taskValues: f.taskValues, flowAssistant: f}
	lv.SetServiceDataByPluginType(common.TypeRateLimiter, windowSet)
	return nil
}

//服务更新回调，找到具体的限流窗口集合，然后触发更新
func (f *FlowQuotaAssistant) OnServiceUpdated(event *common.PluginEvent) error {
	svcEventObject := event.EventObject.(*common.ServiceEventObject)
	if svcEventObject.SvcEventKey.Type != model.EventInstances &&
		svcEventObject.SvcEventKey.Type != model.EventRateLimiting {
		return nil
	}
	newValue := svcEventObject.NewValue.(model.RegistryValue)
	if !newValue.IsInitialized() {
		return nil
	}
	var svcInstances model.ServiceInstances
	if svcEventObject.SvcEventKey.Type == model.EventInstances {
		svcInstances = svcEventObject.NewValue.(model.ServiceInstances)
	} else {
		svcInstances = f.registry.GetInstances(
			&svcEventObject.SvcEventKey.ServiceKey, true, true)
	}
	windowSet := GetRateLimitWindowSet(svcInstances)
	if nil == windowSet {
		return nil
	}
	windowSet.OnServiceUpdated(svcEventObject)
	return nil
}

//服务删除回调
func (f *FlowQuotaAssistant) OnServiceDeleted(event *common.PluginEvent) error {
	svcEventObject := event.EventObject.(*common.ServiceEventObject)
	if svcEventObject.SvcEventKey.Type != model.EventInstances {
		return nil
	}
	svcInstances := svcEventObject.OldValue.(model.ServiceInstances)
	if !svcInstances.IsInitialized() {
		return nil
	}
	windowSet := GetRateLimitWindowSet(svcInstances)
	if nil == windowSet {
		return nil
	}
	windowSet.OnServiceDeleted()
	return nil
}

//获取配额
func (f *FlowQuotaAssistant) GetQuota(commonRequest *data.CommonRateLimitRequest) (*model.QuotaFutureImpl, error) {
	if !f.enable {
		//没有限流规则，直接放通
		resp := &model.QuotaResponse{
			Code: model.QuotaResultOk,
			Info: Disabled,
		}
		return model.NewQuotaFuture(resp, clock.GetClock().Now(), nil), nil
	}
	window, loaded, err := f.lookupRateLimitWindow(commonRequest)
	if nil != err {
		return nil, err
	}
	if nil == window {
		//没有限流规则，直接放通
		resp := &model.QuotaResponse{
			Code: model.QuotaResultOk,
			Info: RuleNotExists,
		}
		gauge := &RateLimitGauge{
			EmptyInstanceGauge: model.EmptyInstanceGauge{},
			Window:             nil,
			Namespace:          commonRequest.DstService.Namespace,
			Service:            commonRequest.DstService.Service,
			Type:               QuotaGranted,
		}
		f.engine.SyncReportStat(model.RateLimitStat, gauge)
		return model.NewQuotaFuture(resp, clock.GetClock().Now(), nil), nil
	}
	if !loaded {
		if err = window.Init(&commonRequest.Criteria, f.rLimitConnector); nil != err {
			return nil, err
		}
	} else if window.GetStatus() == Expired && window.CasStatus(Expired, Created) {
		// 处理超时的window，重新激活
		if err = window.Init(&commonRequest.Criteria, f.rLimitConnector); nil != err {
			return nil, err
		}
	}
	err = window.WaitRemoteInitialized()
	if nil != err {
		return nil, err
	}
	return window.AllocateQuota()
}

//计算限流窗口
func (f *FlowQuotaAssistant) lookupRateLimitWindow(
	commonRequest *data.CommonRateLimitRequest) (*RateLimitWindow, bool, error) {
	for {
		var err error
		// 1. 并发获取被调服务信息和限流配置，服务不存在，返回错误
		if err = f.engine.SyncGetResources(commonRequest); nil != err {
			return nil, true, err
		}
		// 2. 寻找匹配的规则
		rule, err := lookupRule(commonRequest.RateLimitRule, commonRequest.Cluster, commonRequest.Labels)
		if nil != err {
			return nil, true, err
		}
		if nil == rule {
			return nil, true, nil
		}
		commonRequest.Criteria.DstRule = rule
		// 2.获取已有的QuotaWindow
		labelStr := commonRequest.FormatLabelToStr(rule)
		windowSet, window := GetRateLimitWindow(commonRequest.Criteria.DstService, rule,
			labelStr)
		if nil != window {
			//已经存在限流窗口，则直接分配
			return window, true, nil
		}

		//检查是否达到最大限流窗口数量
		nowWindowCount := f.GetWindowCount()
		log.GetBaseLogger().Tracef("RateLimit nowWindowCount:%d %d", nowWindowCount, f.maxWindowSize)
		if nowWindowCount >= f.maxWindowSize {
			count := atomic.LoadUint64(&f.windowCountLogCtrl)
			if count%10000 == 0 {
				log.GetBaseLogger().Infof("RateLimit reach maxWindowSize nowCount:%d maxCount:%d", nowWindowCount, f.maxWindowSize)
			}
			atomic.AddUint64(&f.windowCountLogCtrl, 1)
			return nil, true, nil
		}
		// 3.创建限流窗口
		rateLimitCache := commonRequest.RateLimitRule.GetRuleCache().GetMessageCache(rule).(*pb.RateLimitRuleCache)
		window = NewRateLimitWindow(windowSet, commonRequest.RateLimitRule.GetRevision(),
			rule, rateLimitCache, commonRequest.ControlParam, f.supplier, labelStr)
		window.engine = f.engine
		if rule.GetType() == namingpb.Rule_LOCAL {
			window.configMode = model.ConfigQuotaLocalMode
		} else {
			if f.rateLimitCluster == nil {
				window.configMode = model.ConfigQuotaLocalMode
			} else {
				window.configMode = model.ConfigQuotaGlobalMode
			}
		}
		var retry bool
		var loaded bool
		retry, window, loaded = windowSet.AddWindow(commonRequest.Criteria.DstService.GetRevision(), window)
		f.AddWindowCount()
		if !retry {
			return window, loaded, nil
		}
		//出现数据不一致，重试获取
		time.Sleep(clock.TimeStep())
	}
}

// 寻址规则
func lookupRule(svcRule model.ServiceRule, cluster string, labels map[string]string) (*namingpb.Rule, error) {
	if reflect2.IsNil(svcRule.GetValue()) {
		// 没有配置限流规则
		return nil, nil
	}
	validateErr := svcRule.GetValidateError()
	if nil != validateErr {
		return nil, model.NewSDKError(model.ErrCodeInvalidRule, validateErr,
			"invalid rateLimit rule, please check rule for (namespace=%s, service=%s)",
			svcRule.GetNamespace(), svcRule.GetService())
	}
	ruleCache := svcRule.GetRuleCache()
	rateLimiting := svcRule.GetValue().(*namingpb.RateLimit)
	return matchRuleByLabels(cluster, labels, rateLimiting, ruleCache), nil
}

//通过业务标签来匹配规则
func matchRuleByLabels(cluster string,
	labels map[string]string, ruleSet *namingpb.RateLimit, ruleCache model.RuleCache) *namingpb.Rule {
	if len(ruleSet.Rules) == 0 {
		return nil
	}
	for _, rule := range ruleSet.Rules {
		if len(cluster) > 0 && cluster != rule.GetCluster().GetValue() {
			continue
		}
		if nil != rule.GetDisable() && rule.GetDisable().GetValue() {
			//规则被停用
			continue
		}
		if len(rule.Labels) == 0 {
			//没有业务标签，代表全匹配
			return rule
		}
		var allLabelsMatched = true
		for labelKey, labelValue := range rule.Labels {
			if !matchLabels(labelKey, labelValue, labels, ruleCache) {
				allLabelsMatched = false
				break
			}
		}
		if allLabelsMatched {
			return rule
		}
	}
	return nil
}
