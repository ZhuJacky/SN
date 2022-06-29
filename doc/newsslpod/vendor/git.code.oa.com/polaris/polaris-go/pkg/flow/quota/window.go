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
	"context"
	"git.code.oa.com/polaris/polaris-go/pkg/clock"
	"git.code.oa.com/polaris/polaris-go/pkg/config"
	"git.code.oa.com/polaris/polaris-go/pkg/flow/data"
	"git.code.oa.com/polaris/polaris-go/pkg/log"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	"git.code.oa.com/polaris/polaris-go/pkg/model/pb"
	rlimit "git.code.oa.com/polaris/polaris-go/pkg/model/pb/metric"
	namingpb "git.code.oa.com/polaris/polaris-go/pkg/model/pb/v1"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/common"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/ratelimiter"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/serverconnector"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/modern-go/reflect2"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//限流分配窗口的缓存
type RateLimitWindowSet struct {
	//当前的根规则版本信息
	currentRulesRevision string
	//当前服务版本信息
	currentServiceRevision string
	//更新锁
	updateMutex sync.Mutex
	// 限流窗口列表，key为rule，value为window
	WindowByRule *sync.Map
	//任务列表，用于定时调度
	taskValues model.TaskValues
	//储存FlowQuotaAssistant
	flowAssistant *FlowQuotaAssistant
}

////获取版本号列表
//func getRevisions(svcRule model.ServiceRule) model.HashSet {
//	rLimitValue := svcRule.GetValue()
//	hashSet := model.HashSet{}
//	if reflect2.IsNil(rLimitValue) {
//		return hashSet
//	}
//	rLimit := rLimitValue.(*namingpb.RateLimit)
//	for _, rule := range rLimit.GetRules() {
//		hashSet.Add(rule.GetRevision().GetValue())
//	}
//	return hashSet
//}
//
////获取前后已经出现更新的规则版本号
//func calcUpdatedRules(svcEventObject *common.ServiceEventObject) []string {
//	oldSvcRule := svcEventObject.OldValue.(model.ServiceRule)
//	newSvcRule := svcEventObject.NewValue.(model.ServiceRule)
//	oldRevisions := getRevisions(oldSvcRule)
//	newRevisions := getRevisions(newSvcRule)
//	updatedRules := make([]string, 0, len(oldRevisions))
//	for revision := range oldRevisions {
//		if !newRevisions.Contains(revision) {
//			updatedRules = append(updatedRules, revision.(string))
//		}
//	}
//	return updatedRules
//}

//服务更新回调
func (rs *RateLimitWindowSet) OnServiceDeleted() {
	//清理map中所有已经调度的任务
	rs.updateMutex.Lock()
	defer rs.updateMutex.Unlock()
	rs.WindowByRule.Range(func(key, value interface{}) bool {
		window := value.(*RateLimitWindow)
		window.SetStatus(Deleted)
		rs.taskValues.DeleteValue(key.(string), window)
		return true
	})
}

//服务更新回调
func (rs *RateLimitWindowSet) OnServiceUpdated(svcEventObject *common.ServiceEventObject) {
	var updatedRules *common.RateLimitDiffInfo
	if svcEventObject.SvcEventKey.Type == model.EventRateLimiting {
		updatedRules = svcEventObject.DiffInfo.(*common.RateLimitDiffInfo)
	}
	rs.updateMutex.Lock()
	defer rs.updateMutex.Unlock()
	switch svcEventObject.SvcEventKey.Type {
	case model.EventInstances:
		svcInstances := svcEventObject.NewValue.(model.ServiceInstances)
		if !svcInstances.IsInitialized() {
			return
		}
		count := 0
		for _, inst := range svcInstances.GetInstances() {
			if !inst.IsHealthy() || inst.IsIsolated() {
				continue
			}
			count++
		}
		if count == 0 {
			count = 1
		}
		//服务实例产生变更，则更新服务实例个数
		rs.WindowByRule.Range(func(key, value interface{}) bool {
			value.(*RateLimitWindow).OnInstancesChanged(count)
			return true
		})
		rs.currentServiceRevision = svcInstances.GetRevision()
	case model.EventRateLimiting:
		// 比较新旧变化，进行剔除
		// 可能有正则区分统计，考虑到更新不频繁，数量级不会太大，简单粗暴处理
		for _, revision := range updatedRules.DeletedRules {
			deleteRsSpreadWindow(rs, revision)
		}
		for _, revisionChange := range updatedRules.UpdatedRules {
			deleteRsSpreadWindow(rs, revisionChange.OldRevision)
		}
		rs.currentRulesRevision = svcEventObject.NewValue.(model.ServiceRule).GetRevision()
	}
}

func deleteRsSpreadWindow(rs *RateLimitWindowSet, revision string) {
	rs.WindowByRule.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		if strings.Contains(keyStr, revision) {
			deleteWindow(rs, key.(string))
		}
		return true
	})
}

//从RateLimitWindowSet中删除一个RateLimitWindow
func deleteWindow(rs *RateLimitWindowSet, key string) {
	value, ok := rs.WindowByRule.Load(key)
	if ok {
		window := value.(*RateLimitWindow)
		window.SetStatus(Deleted)
		// 进行一次上报
		rs.WindowByRule.Delete(key)
		rs.taskValues.DeleteValue(key, window)
		rs.flowAssistant.DelWindowCount()
		//旧有的窗口被删除了，那么进行一次上报
		window.engine.SyncReportStat(model.RateLimitStat, &RateLimitGauge{
			EmptyInstanceGauge: model.EmptyInstanceGauge{},
			Window:             window,
			Type:               WindowDeleted,
		})
	}
}

//插入限流窗口
func (rs *RateLimitWindowSet) AddWindow(
	svcRevision string, window *RateLimitWindow) (retry bool, curWindow *RateLimitWindow, loaded bool) {
	rs.updateMutex.Lock()
	defer rs.updateMutex.Unlock()
	rulesRevision := window.rulesRevision
	//出现了变更，数据不一致，前端重试
	if len(rs.currentServiceRevision) > 0 && rs.currentServiceRevision != svcRevision {
		return true, nil, false
	}
	if len(rs.currentRulesRevision) > 0 && rs.currentRulesRevision != rulesRevision {
		return true, nil, false
	}
	//数据一致，则直接设置
	key := window.WindowSetKey
	curWindowValue, loaded := rs.WindowByRule.LoadOrStore(key, window)
	return false, curWindowValue.(*RateLimitWindow), loaded
}

//删除窗口
func (rs *RateLimitWindowSet) DeleteWindow(window *RateLimitWindow) {
	rs.updateMutex.Lock()
	defer rs.updateMutex.Unlock()
	rs.WindowByRule.Delete(window.WindowSetKey)
}

func (rs *RateLimitWindowSet) DeleteWindowWithSyncReport(window *RateLimitWindow) {
	rs.updateMutex.Lock()
	defer rs.updateMutex.Unlock()
	rs.WindowByRule.Delete(window.WindowSetKey)
}

//获取配额分配窗口集合
func GetRateLimitWindowSet(svcInstances model.ServiceInstances) *RateLimitWindowSet {
	value := svcInstances.(*pb.ServiceInstancesInProto).
		GetServiceLocalValue().GetServiceDataByPluginType(common.TypeRateLimiter)
	var windowSet *RateLimitWindowSet
	if !reflect2.IsNil(value) {
		windowSet = value.(*RateLimitWindowSet)
	}
	return windowSet
}

//获取配额分配窗口
func GetRateLimitWindow(
	svcInstances model.ServiceInstances, rule *namingpb.Rule, label string) (*RateLimitWindowSet, *RateLimitWindow) {
	windowSet := GetRateLimitWindowSet(svcInstances)
	var key string
	if rule.GetRegexCombine() != nil && rule.GetRegexCombine().Value {
		key = rule.GetRevision().GetValue()
	} else {
		key = rule.GetRevision().GetValue() + config.DefaultNamesSeparator + label
	}
	window, ok := windowSet.WindowByRule.Load(key)
	if ok {
		return windowSet, window.(*RateLimitWindow)
	}
	return windowSet, nil
}

const (
	//刚创建， 无需进行后台调度
	Created int32 = iota
	//已获取调度权，准备开始调度
	Initializing
	//已经在远程初始化结束
	Initialized
	//从远程拉取配额成功
	Acquired
	//远程初始化失败
	RemoteInitFail
	//远程更新配额失败
	RemoteAcquireFail
	//太久没有人访问，窗口过期，重新激活后需要重新init
	Expired
	//任务删除，无效规则或者已经被移除的规则
	Deleted
	//
	NeedRateLimitInit
)

const (
	//定时轮询模式，默认模式
	SyncInterval int32 = iota
	//立刻上报模式，当配额使用达到80%时，会转换成该模式，此时下一次调度立刻上报（无论是否到时间）
	SyncAtOnce
)

// 远程同步相关参数
type RemoteSyncParam struct {
	// 连接相关参数
	model.ControlParam
	// 是否已经完成远程初始化
	remoteInitCtx atomic.Value
	//上报周期
	reportInterval time.Duration
	//远程与本地之间的时间差, 毫秒，int64
	remoteTimeDiff atomic.Value
	//调度标识，是否立刻执行一次远程上报
	syncFlag int32
}

// 配额使用信息
type UsageInfo struct {
	//配额使用时间
	CurTime int64
	//配额使用详情
	QuotaUsed map[int64]uint32

	Limited map[int64]uint32
}

// 远程下发配额
type RemoteQuotaResult struct {
	// 上报时的使用信息，用于成功后扣除
	CurrentUsage *UsageInfo
	// 远程配额查询结果
	RemoteQuotas []*rlimit.Limiter
	// 远程服务器时间
	ServerTime int64
}

// 远程配额分配的令牌桶
type RemoteAwareBucket interface {
	// 父接口，执行用户配额分配操作
	model.QuotaAllocator
	//设置通过限流服务端获取的远程配额
	SetRemoteQuota(*RemoteQuotaResult)
	// 获取已经分配的配额
	GetQuotaUsed(nowTime int64) *UsageInfo
	// 获取已经分配的配额(调用后清零)
	GetQuotaUsedForAcquire(nowTime int64) *UsageInfo
	// 更新服务实例数量
	UpdateInstanceCount(count int32)
	// 初始化sliceWindow周期开始时间（和server开始窗口对齐）
	InitPeriodStart(now int64)
	//GetMaxRemoteWait 获取最小周期
	GetMaxRemoteWait() int64
}

// 限流窗口
type RateLimitWindow struct {
	//配额窗口集合
	WindowSet *RateLimitWindowSet
	// 根规则的版本信息
	rulesRevision string
	// 服务信息
	SvcKey model.ServiceKey
	// 主键
	quotaKey string
	// 已经匹配到的限流规则，没有匹配则为空
	// 由于可能会出现规则并没有发生变化，但是缓存对象更新的情况，因此这里使用原子变量
	Rule *namingpb.Rule
	// 正则对应的label
	labels string
	// 存储在windowSet中的key
	WindowSetKey string
	//淘汰周期，取最大统计周期+1s
	expireDuration time.Duration
	// 远程同步参数
	syncParam RemoteSyncParam
	// 流量整形算法桶
	trafficShapingBucket ratelimiter.QuotaBucket
	// 执行正式分配的令牌桶
	allocatingBucket RemoteAwareBucket
	// 限流插件
	rateLimiter ratelimiter.ServiceRateLimiter
	// 窗口状态
	status int32
	//// 用于保持老接口init， 中间临时方案，完全迁移到metric-server后删除
	//oldRateLimitInitStatus int32
	// 用于保持新metric server接口init， 中间临时方案，完全迁移到metric-server后删除
	metricServerInitStatus int32
	//最后一次获取限流配额时间
	lastQuotaAccessTime atomic.Value
	//其他插件在这里添加的相关数据，一般是统计插件使用
	PluginData map[int32]interface{}
	//对应sdkcontext的flow engine
	engine model.Engine
	// metric（1s周期）上报，用于监控
	//statisticsSlice []*StatisticsBucket
	// 单位 秒
	lastReportTime int64
	//最近一次拉取远程配额返回的时间,单位ms
	lastRemoteDealQuotaTime int64
	//最近一次发送acquire远程同步配额的时间, 单位ns
	lastAcquireTime int64
	//表示同步配额的请求是否收到回包
	acquireNotFinish int32

	//初始化后指定的限流模式（本地或远程）
	configMode model.ConfigMode

	//window的初始化时间, 单位ns
	windowInitTime int64

	//连续同步远程配额acquire, 收到回包很慢的计算
	continuousSlowRemoteAcquireCount int32
}

//超过多长时间后进行淘汰，淘汰后需要重新init
var (
	// 淘汰因子，过期时间=MaxDuration + ExpireFactor
	ExpireFactor = 3 * time.Second

	DefaultStatisticReportPeriod = 1 * time.Second
)

//计算淘汰周期
func getExpireDuration(maxDuration time.Duration) time.Duration {
	expireDuration := maxDuration + ExpireFactor
	return expireDuration
}

// 创建限流窗口
func NewRateLimitWindow(windowSet *RateLimitWindowSet, rulesRevision string,
	rule *namingpb.Rule, rateLimitCache *pb.RateLimitRuleCache,
	controlParam model.ControlParam, supplier plugin.Supplier, label string) *RateLimitWindow {
	window := &RateLimitWindow{}
	window.WindowSet = windowSet
	window.rulesRevision = rulesRevision
	window.quotaKey = buildQuotaKey(rule, label)
	window.SvcKey.Service = rule.GetService().GetValue()
	window.SvcKey.Namespace = rule.GetNamespace().GetValue()
	window.syncParam.ControlParam = controlParam
	window.syncParam.syncFlag = SyncInterval
	window.status = Created
	window.Rule = rule
	window.PluginData = make(map[int32]interface{})
	// 初始化规则
	if rule.GetReport().GetInterval() != nil {
		window.syncParam.reportInterval, _ = pb.ConvertDuration(rule.GetReport().GetInterval())
		if window.syncParam.reportInterval < config.MinRateLimitReportInterval {
			window.syncParam.reportInterval = config.MinRateLimitReportInterval
		}
	} else {
		window.syncParam.reportInterval = config.DefaultRateLimitAcquireInterval
	}

	window.expireDuration = getExpireDuration(rateLimitCache.MaxDuration)
	window.rateLimiter = createBehavior(supplier, rule.GetAction().GetValue())
	window.lastQuotaAccessTime.Store(clock.GetClock().Now())
	window.lastAcquireTime = 0
	window.acquireNotFinish = 0
	//创建对应
	handlers := supplier.GetEventSubscribers(common.OnRateLimitWindowCreated)
	if len(handlers) > 0 {
		eventObj := &common.PluginEvent{
			EventType:   common.OnRateLimitWindowCreated,
			EventObject: window,
		}
		for _, h := range handlers {
			h.Callback(eventObj)
		}
	}
	window.labels = label
	if rule.GetRegexCombine() != nil && rule.GetRegexCombine().Value {
		window.WindowSetKey = window.Rule.GetRevision().Value
	} else {
		window.WindowSetKey = window.Rule.GetRevision().Value + config.DefaultNamesSeparator + label
	}
	window.lastReportTime = clock.GetClock().Now().Unix()
	window.metricServerInitStatus = Initializing
	return window
}

//获取quotaKey
func buildQuotaKey(rule *namingpb.Rule, label string) string {
	//<规则的ID>#<subset>#<metadata>
	builder := &strings.Builder{}
	builder.Grow(len(rule.GetId().GetValue()))
	builder.WriteString(rule.GetId().GetValue())
	builder.WriteString(config.DefaultNamesSeparator)
	builder.WriteString(config.DefaultNamesSeparator)
	if rule.GetRegexCombine() != nil && rule.GetRegexCombine().Value {
		ruleLabels := formatRuleLabelsToStr(rule)
		builder.WriteString(ruleLabels)
	} else {
		builder.WriteString(label)
	}
	return builder.String()
}

//根据限流行为名获取限流算法插件
func createBehavior(supplier plugin.Supplier, behaviorName string) ratelimiter.ServiceRateLimiter {
	//因为构造缓存时候已经校验过，所以这里可以直接忽略错误
	plug, _ := supplier.GetPlugin(common.TypeRateLimiter, behaviorName)
	return plug.(ratelimiter.ServiceRateLimiter)
}

//校验输入的元数据是否符合规则
func matchLabels(ruleMetaKey string, ruleMetaValue *namingpb.MatchString,
	labels map[string]string, ruleCache model.RuleCache) bool {
	if len(labels) == 0 {
		return false
	}
	var value string
	var ok bool
	if value, ok = labels[ruleMetaKey]; !ok {
		//集成的路由规则不包含这个key，就不匹配
		return false
	}
	ruleMetaValueStr := ruleMetaValue.GetValue().GetValue()
	switch ruleMetaValue.Type {
	case namingpb.MatchString_REGEX:
		regexObj := ruleCache.GetRegexMatcher(ruleMetaValueStr)
		if !regexObj.MatchString(value) {
			return false
		}
		return true
	default:
		return value == ruleMetaValueStr
	}
}

//上下文的键类型
type contextKey struct {
	name string
}

//ToString方法
func (k *contextKey) String() string { return "rateLimit context value " + k.name }

//key，用于共享错误信息
var errKey = &contextKey{name: "ctxError"}

//错误容器，用于传递上下文错误信息
type errContainer struct {
	err atomic.Value
}

// 初始化限流窗口
func (r *RateLimitWindow) Init(
	criteria *ratelimiter.InitCriteria, connector serverconnector.RateLimitConnector) error {
	if r.GetStatus() != Created {
		return nil
	}
	container := &errContainer{}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), errKey, container))
	defer cancel()
	r.syncParam.remoteInitCtx.Store(ctx)
	r.SetStatus(Initializing)
	// 1. 初始化流量整形窗口
	bucket, err := r.rateLimiter.InitQuota(criteria)

	if nil != err {
		log.GetBaseLogger().Errorf("fail to call InitQuota by plugin %s, err is %v", r.rateLimiter.Name(), err)
		container.err.Store(err)
		r.SetStatus(Deleted)
		r.WindowSet.DeleteWindow(r)
		return err
	}
	r.acquireNotFinish = 0
	r.windowInitTime = time.Now().UnixNano()
	//加入定时轮询
	if r.configMode == model.ConfigQuotaGlobalMode {
		r.WindowSet.taskValues.AddValue(r.WindowSetKey, r)
	}
	r.trafficShapingBucket = bucket
	// 2. 初始化配额窗口
	if r.Rule.GetResource() == namingpb.Rule_QPS {
		r.allocatingBucket = NewRemoteAwareQpsBucket(r, r.Rule, int32(len(criteria.DstService.GetInstances())))
	}
	// 3. 向server初始化任务
	if r.configMode == model.ConfigQuotaGlobalMode {
		go r.doRemoteInitialize(connector)
	}
	return nil
}

func (r *RateLimitWindow) GetRulesRevision() string {
	return r.rulesRevision
}

func (r *RateLimitWindow) GetWindowInitTime() int64 {
	return r.windowInitTime
}

//转换成限流PB初始化消息
func (r *RateLimitWindow) InitializeRequest() *rlimit.RateLimitRequest {
	rlReq := &rlimit.RateLimitRequest{}
	rlReq.Key = &wrappers.StringValue{Value: r.quotaKey}
	rlReq.Namespace = &wrappers.StringValue{Value: r.SvcKey.Namespace}
	rlReq.Service = &wrappers.StringValue{Value: r.SvcKey.Service}
	rule := r.Rule
	rlReq.Totals = make([]*rlimit.Limiter, 0, len(rule.Amounts))
	for _, amount := range rule.Amounts {
		rlReq.Totals = append(rlReq.Totals, &rlimit.Limiter{
			Amount:   amount.MaxAmount,
			Duration: amount.ValidDuration,
		})
	}
	return rlReq
}

// 与限流server的应答包装器，带上ttl
type responseWrapper struct {
	// 应答消息
	resp proto.Message
	// 消息的TTL，单位毫秒
	ttl int64
}

// 远程访问的错误信息
type RemoteErrorContainer struct {
	sdkErr atomic.Value
}

// 发起远程初始化
func (r *RateLimitWindow) doRemoteInitialize(connector serverconnector.RateLimitConnector) {
	log.GetBaseLogger().Debugf("[RateLimit]start to doRemoteInitialize for service %s", r.SvcKey)
	if r.Rule.GetType() == namingpb.Rule_LOCAL || r.configMode == model.ConfigQuotaLocalMode {
		return
	}
	var wrapper *responseWrapper
	param := r.syncParam.ControlParam
	initRequest := r.InitializeRequest()
	respMsg, sdkErr := data.RetrySyncCall("rateLimitInit", &r.SvcKey, initRequest,
		func(request interface{}) (interface{}, error) {
			startTime := time.Now()
			resp, err := connector.Initialize(request.(proto.Message), param.Timeout)
			if nil != err {
				return nil, err
			}
			duration := model.ToMilliSeconds(time.Since(startTime))
			return &responseWrapper{
				resp: resp,
				ttl:  duration / 2,
			}, nil
		}, &param)
	if nil != sdkErr {
		log.GetBaseLogger().Errorf(
			"fail to call RateLimitService.Initialize, service %s, quotaKey %s, error is %s",
			r.SvcKey, r.quotaKey, sdkErr)
		//远端server不可用，初始化失败，使用本地模式
		r.SetStatus(RemoteInitFail)
		return
	}
	wrapper = respMsg.(*responseWrapper)
	r.OnResponse(nil, wrapper.resp.(*rlimit.RateLimitResponse), wrapper.ttl, true)
	log.GetBaseLogger().Infof("doRemoteInitialize success %s req:%s", r.SvcKey.String(),
		initRequest.GetKey().GetValue())
	r.SetStatus(Initialized)
	realResp := wrapper.resp.(*rlimit.RateLimitResponse)
	r.allocatingBucket.InitPeriodStart(realResp.GetTimestamp().GetValue())
}

//// 组装MetricInitRequest
//func (r *RateLimitWindow) metricInitRequest() *rlimit.MetricInitRequest  {
//	initReq := rlimit.MetricInitRequest{
//		Key:        &rlimit.MetricKey{
//			Service:   r.SvcKey.Service,
//			Namespace: r.SvcKey.Namespace,
//			Subset:    "",
//			Labels:    r.labels,
//			Role:      rlimit.MetricKey_Callee,
//		},
//	}
//	totalDim := rlimit.MetricDimension{
//		Type:  rlimit.MetricType_ReqCount,
//		Value: "",
//	}
//	initReq.Dimensions = append(initReq.Dimensions, &totalDim)
//
//	limitDim := rlimit.MetricDimension{
//		Type:  rlimit.MetricType_LimitCount,
//		Value: "",
//	}
//	initReq.Dimensions = append(initReq.Dimensions, &limitDim)
//
//	// 临时方案，Duration Precision 先按report情况写死
//	win := rlimit.MetricInitRequest_MetricWindow{
//		Duration:  1000,
//		Precision: 1,
//	}
//	initReq.Windows = append(initReq.Windows, &win)
//
//	return &initReq
//}

//// 完全迁移到metric server后，删除（统一使用doRemoteInitialize）
//func (r *RateLimitWindow) doMetricReportInit(connector serverconnector.RateLimitConnector)  {
//	if r.metricServerInitStatus == Initialized {
//		return
//	}
//	var wrapper *responseWrapper
//	param := r.syncParam.ControlParam
//	initReq := r.metricInitRequest()
//	respMsg, sdkErr := data.RetrySyncCall("rateLimitInit", &r.SvcKey, initReq,
//		func(request interface{}) (interface{}, error) {
//			startTime := time.Now()
//			resp, err := connector.Init(request.(proto.Message), param.Timeout)
//			if nil != err {
//				return nil, err
//			}
//			duration := model.ToMilliSeconds(time.Since(startTime))
//			return &responseWrapper{
//				resp: resp,
//				ttl:  duration / 2,
//			}, nil
//		}, &param)
//	if nil != sdkErr {
//		log.GetBaseLogger().Errorf(
//			"fail to call doMetricReportInit, service %s, quotaKey %s, error is %s",
//			r.SvcKey, r.quotaKey, sdkErr)
//		return
//	} else {
//		log.GetBaseLogger().Tracef("RateLimitService.doMetricReportInit done  service %s ", r.SvcKey)
//	}
//	wrapper = respMsg.(*responseWrapper)
//	r.OnMetricInitResponse(wrapper.resp.(*rlimit.MetricResponse), wrapper.ttl)
//	r.metricServerInitStatus = Initialized
//}

//比较两个窗口是否相同
func (r *RateLimitWindow) CompareTo(another interface{}) int {
	return strings.Compare(r.quotaKey, another.(*RateLimitWindow).quotaKey)
}

//删除前进行检查，返回true才删除，该检查是同步操作
func (r *RateLimitWindow) EnsureDeleted(value interface{}) bool {
	//只有过期才删除
	return r.GetStatus() == Expired || r.GetStatus() == Deleted
}

//转换成限流PB上报消息
func (r *RateLimitWindow) acquireRequest() (*UsageInfo, *rlimit.RateLimitRequest) {
	rlReq := &rlimit.RateLimitRequest{}
	rlReq.Key = &wrappers.StringValue{Value: r.quotaKey}
	rlReq.Namespace = &wrappers.StringValue{Value: r.SvcKey.Namespace}
	rlReq.Service = &wrappers.StringValue{Value: r.SvcKey.Service}
	rule := r.Rule
	curTimeMilli := time.Now().UnixNano() / 1e6
	serverTimeMilli := curTimeMilli + r.GetRemoteDiff()
	rlReq.Timestamp = &wrappers.Int64Value{
		Value: serverTimeMilli,
	}
	usedQuotas := r.allocatingBucket.GetQuotaUsedForAcquire(serverTimeMilli)
	rlReq.Useds = make([]*rlimit.Limiter, 0, len(rule.Amounts))
	for _, amount := range rule.Amounts {
		duration := amount.ValidDuration
		goDuration, _ := pb.ConvertDuration(duration)
		rlReq.Totals = append(rlReq.Totals, &rlimit.Limiter{
			Amount: &wrappers.UInt32Value{
				Value: amount.GetMaxAmount().GetValue(),
			},
			Duration: amount.ValidDuration,
		})
		quotaUsed := usedQuotas.QuotaUsed[model.ToMilliSeconds(goDuration)]
		limited := usedQuotas.Limited[model.ToMilliSeconds(goDuration)]
		rlReq.Useds = append(rlReq.Useds, &rlimit.Limiter{
			Amount: &wrappers.UInt32Value{
				Value: quotaUsed,
			},
			Duration: amount.ValidDuration,
			Limited:  &wrappers.UInt32Value{Value: limited},
		})
	}
	return usedQuotas, rlReq
}

func (r *RateLimitWindow) assembleAcquireRequestOnlyWithReport(curTime int64, rpDur time.Duration,
	limited uint32) *rlimit.RateLimitRequest {
	rlReq := &rlimit.RateLimitRequest{}
	rlReq.Key = &wrappers.StringValue{Value: r.quotaKey}
	rlReq.Namespace = &wrappers.StringValue{Value: r.SvcKey.Namespace}
	rlReq.Service = &wrappers.StringValue{Value: r.SvcKey.Service}
	rule := r.Rule
	rlReq.Timestamp = &wrappers.Int64Value{
		Value: curTime,
	}
	rlReq.Useds = make([]*rlimit.Limiter, 0, len(rule.Amounts))
	for _, amount := range rule.Amounts {
		duration := amount.ValidDuration
		goDuration, _ := pb.ConvertDuration(duration)
		if goDuration != rpDur {
			continue
		}
		rlReq.Useds = append(rlReq.Useds, &rlimit.Limiter{
			Amount: &wrappers.UInt32Value{
				Value: 0,
			},
			Duration: amount.ValidDuration,
			Limited:  &wrappers.UInt32Value{Value: limited},
		})
	}
	return rlReq
}

//执行限流上报
//func (r *RateLimitWindow) doRemoteAcquire(connector serverconnector.RateLimitConnector) {
//	param := r.syncParam.ControlParam
//	log.GetBaseLogger().Tracef("start to doRemoteAcquire for service %s, Rule %s\n", r.SvcKey, r.Rule.GetId())
//
//	_, request := r.acquireRequest()
//	//log.GetBaseLogger().Infof("----------SetRemoteQuota doRemoteAcquire req:%s\n", request.String())
//	respMsg, sdkErr := data.RetrySyncCall("rateLimitAcquire", &r.SvcKey, request,
//		func(request interface{}) (interface{}, error) {
//			startTime := time.Now()
//			resp, err := connector.Acquire(request.(proto.Message), param.Timeout)
//			if nil != err {
//				return nil, err
//			}
//			duration := model.ToMilliSeconds(time.Since(startTime))
//			return &responseWrapper{
//				resp: resp,
//				ttl:  duration / 2,
//			}, nil
//		}, &param)
//	if nil != sdkErr {
//		log.GetBaseLogger().Errorf(
//			"fail to call RateLimitService.Acquire, service %s, quotaKey %s, error is %s",
//			r.SvcKey, r.quotaKey, sdkErr)
//		r.SetStatus(RemoteAcquireFail)
//		return
//	}
//	wrapper := respMsg.(*responseWrapper)
//	timeNowMill := model.ParseMilliSeconds(clock.GetClock().Now().UnixNano())
//	usedQuota := r.allocatingBucket.GetQuotaUsed(timeNowMill)
//	r.OnResponse(usedQuota, wrapper.resp.(*rlimit.RateLimitResponse), wrapper.ttl, true)
//	r.SetStatus(Acquired)
//}

//处理应答对象
func (r *RateLimitWindow) OnResponse(usedQuota *UsageInfo, resp *rlimit.RateLimitResponse, ttl int64, setDiff bool) {
	curTimeMilli := time.Now().UnixNano() / 1e6
	serverTimeMilli := resp.GetTimestamp().GetValue()
	if setDiff {
		var diff int64
		if serverTimeMilli > 0 {
			diff = serverTimeMilli + ttl - curTimeMilli
		}
		r.syncParam.remoteTimeDiff.Store(diff)
	}
	r.allocatingBucket.SetRemoteQuota(&RemoteQuotaResult{
		CurrentUsage: usedQuota,
		RemoteQuotas: resp.GetSumUseds(),
		ServerTime:   resp.GetTimestamp().GetValue(),
	})
}

// 原子获取状态
func (r *RateLimitWindow) GetStatus() int32 {
	return atomic.LoadInt32(&r.status)
}

// 设置状态
func (r *RateLimitWindow) SetStatus(status int32) {
	atomic.StoreInt32(&r.status, status)
}

// CAS设置状态
func (r *RateLimitWindow) CasStatus(oldStatus int32, status int32) bool {
	return atomic.CompareAndSwapInt32(&r.status, oldStatus, status)
}

// 等待远程初始化结束
func (r *RateLimitWindow) WaitRemoteInitialized() error {
	status := r.GetStatus()
	for status == Created {
		//CAS操作，保证原子变量ctx可以正常设置
		status = r.GetStatus()
	}
	var ctx context.Context
	ctx = r.syncParam.remoteInitCtx.Load().(context.Context)
	if status != Initializing {
		goto finally
	}
	<-ctx.Done()
finally:
	if status != Deleted {
		return nil
	}
	container := ctx.Value(errKey).(*errContainer)
	return container.err.Load().(error)
}

// 获取远程校正时间
func (r *RateLimitWindow) GetRemoteDiff() int64 {
	if remoteDiffValue := r.syncParam.remoteTimeDiff.Load(); !reflect2.IsNil(remoteDiffValue) {
		return remoteDiffValue.(int64)
	}
	return 0
}

// 分配配额
func (r *RateLimitWindow) AllocateQuota() (*model.QuotaFutureImpl, error) {
	now := clock.GetClock().Now()
	r.lastQuotaAccessTime.Store(now)
	shapingResult, err := r.trafficShapingBucket.GetQuota()
	if nil != err {
		return nil, err
	}
	totalGauge := &RateLimitGauge{
		EmptyInstanceGauge: model.EmptyInstanceGauge{},
		Window:             r,
		Type:               QuotaRequested,
	}
	r.engine.SyncReportStat(model.RateLimitStat, totalGauge)
	if shapingResult.Code == model.QuotaResultLimited {
		//如果结果是拒绝了分配，那么进行一次上报
		gauge := &RateLimitGauge{
			EmptyInstanceGauge: model.EmptyInstanceGauge{},
			Window:             r,
			Type:               TrafficShapingLimited,
		}
		r.engine.SyncReportStat(model.RateLimitStat, gauge)

		resp := &model.QuotaResponse{
			Code: model.QuotaResultLimited,
			Info: shapingResult.Info,
		}
		return model.NewQuotaFuture(resp, now, nil), nil
	}
	deadline := now
	if shapingResult.QueueTime > 0 {
		factor := math.Floor(float64(shapingResult.QueueTime) / float64(clock.TimeStep()))
		deadline = now.Add(clock.TimeStep() * time.Duration(factor))
	}
	return model.NewQuotaFuture(nil, deadline, r.allocatingBucket), nil
}

//获取最近访问时间
func (r *RateLimitWindow) GetLastQuotaAccessTime() time.Time {
	return r.lastQuotaAccessTime.Load().(time.Time)
}

//实例变更回调
func (r *RateLimitWindow) OnInstancesChanged(instCount int) {
	r.trafficShapingBucket.OnInstancesChanged(instCount)
	r.allocatingBucket.UpdateInstanceCount(int32(instCount))
}

// metric report response
func (r *RateLimitWindow) OnMetricResponse(resp *rlimit.MetricResponse, ttl int64) {
	if log.GetBaseLogger().IsLevelEnabled(log.TraceLog) {
		log.GetBaseLogger().Tracef("OnMetricResponse done msgId:%d ttl:%d", resp.GetMsgId().Value, ttl)
	}
}

// metric init response
func (r *RateLimitWindow) OnMetricInitResponse(resp *rlimit.MetricResponse, ttl int64) {
	if log.GetBaseLogger().IsLevelEnabled(log.TraceLog) {
		log.GetBaseLogger().Tracef("OnMetricResponse done msgId:%d ttl:%d", resp.GetMsgId().Value, ttl)
	}
}

// metric report限流统计上报
//func (r *RateLimitWindow) doReport(connector serverconnector.RateLimitConnector) {
//	if r.metricServerInitStatus != Initialized {
//		r.doMetricReportInit(connector)
//		return
//	}
//	size := len(r.statisticsSlice)
//	timeNow := clock.GetClock().Now()
//	timeNowUnix := timeNow.Unix()
//	if timeNow.UnixNano() - r.lastReportTime < int64(DefaultStatisticReportPeriod) {
//		return
//	}
//	param := r.syncParam.ControlParam
//	log.GetBaseLogger().Tracef("start to doReport for service %s, Rule %s\n", r.SvcKey, r.Rule.GetId())
//
//	nowIdx := timeNowUnix % int64(size)
//	var reportDataList []*ReportElements
//	var i int64 = 0
//	for i=0; i<int64(size); i++ {
//		idx := nowIdx - i
//		if idx < 0 {
//			idx += int64(size)
//		}
//		rData := r.statisticsSlice[idx].GetReportData(timeNowUnix - i)
//		reportDataList = append(reportDataList, rData)
//	}
//	request := r.reportRequest(timeNow, reportDataList)
//	respMsg, sdkErr := data.RetrySyncCall("rateLimitReport", &r.SvcKey, request,
//		func(request interface{}) (interface{}, error) {
//			startTime := time.Now()
//			resp, err := connector.Report(request.(proto.Message), param.Timeout)
//			if nil != err {
//				return &responseWrapper{
//					resp: resp,
//					ttl: 0,
//				}, err
//			}
//			duration := model.ToMilliSeconds(time.Since(startTime))
//			return &responseWrapper{
//				resp: resp,
//				ttl:  duration / 2,
//			}, nil
//		}, &param)
//	if nil != sdkErr {
//		log.GetBaseLogger().Errorf(
//			"fail to call RateLimitService.doReport, service %s, quotaKey %s, error is %s",
//			r.SvcKey, r.quotaKey, sdkErr)
//		r.SetStatus(RemoteAcquireFail)
//		r.lastReportTime = timeNow.UnixNano()
//		wrapper := respMsg.(*responseWrapper)
//		if wrapper.resp != nil {
//			rsp := wrapper.resp.(*rlimit.MetricResponse)
//			if rsp.GetCode().Value/1000 == 404 {
//				r.doMetricReportInit(connector)
//			}
//		}
//		return
//	} else {
//		log.GetBaseLogger().Tracef("RateLimitService.doReport done  service %s ", r.SvcKey)
//	}
//	wrapper := respMsg.(*responseWrapper)
//	r.OnMetricResponse(wrapper.resp.(*rlimit.MetricResponse), wrapper.ttl)
//	r.lastReportTime = timeNow.UnixNano()
//}

//// metric report组装请求
//func (r *RateLimitWindow) reportRequest(timeNow time.Time, dataSlice []*ReportElements) *rlimit.MetricRequest {
//	rReq := &rlimit.MetricRequest{}
//	rReq.Key = &rlimit.MetricKey{
//		Service:   r.SvcKey.Service,
//		Namespace: r.SvcKey.Namespace,
//		Subset:    "",
//		Labels:    r.labels,
//		Role:      rlimit.MetricKey_Callee,
//	}
//	// 临时策略，先写死
//	incr := &rlimit.MetricRequest_MetricIncrement{
//		Duration:  1000,
//		Precision: 1,
//	}
//	totalValue := &rlimit.MetricRequest_MetricIncrement_Values{
//		Dimension: &rlimit.MetricDimension{
//			Type:   rlimit.MetricType_ReqCount,
//			Value: "",
//		},
//	}
//	limitValue := &rlimit.MetricRequest_MetricIncrement_Values{
//		Dimension: &rlimit.MetricDimension{
//			Type:   rlimit.MetricType_LimitCount,
//			Value: "",
//		},
//	}
//	for _, v := range dataSlice {
//		totalValue.Values = append(totalValue.Values, v.TotalCount)
//		limitValue.Values = append(limitValue.Values, v.LimitCount)
//	}
//
//	serverTime := time.Now().UnixNano() + r.GetRemoteDiff() * 1000
//	rReq.Timestamp = &wrappers.Int64Value{Value: serverTime}
//	incr.Values = append(incr.Values, totalValue)
//	incr.Values = append(incr.Values, limitValue)
//	rReq.Increments = append(rReq.Increments, incr)
//	msgId := atomic.AddInt64(&MetricReportGlobalMsgId, 1)
//	rReq.MsgId = &wrappers.Int64Value{Value: msgId}
//	return rReq
//}

func formatRuleLabelsToStr(rule *namingpb.Rule) string {
	if len(rule.GetLabels()) == 0 {
		return ""
	}
	var tmpList []string
	for k, v := range rule.GetLabels() {
		tmpList = append(tmpList, k+config.DefaultMapKeyValueSeparator+v.GetValue().Value)
	}
	sort.Strings(tmpList)
	s := strings.Join(tmpList, config.DefaultMapKVTupleSeparator)
	return s
}
