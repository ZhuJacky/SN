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

package data

import (
	"git.code.oa.com/polaris/polaris-go/pkg/config"
	"git.code.oa.com/polaris/polaris-go/pkg/log"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	namingpb "git.code.oa.com/polaris/polaris-go/pkg/model/pb/v1"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/loadbalancer"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/ratelimiter"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/servicerouter"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	//缓存查询请求的对象池
	instanceRequestPool = &sync.Pool{}
	//缓存规则查询请求的对象池
	ruleRequestPool = &sync.Pool{}
	//限流请求对象池
	rateLimitRequestPool = &sync.Pool{}
	//调用结果上报请求对象池
	serviceCallResultRequestPool = &sync.Pool{}
)

//通过池子获取请求对象
func PoolGetCommonInstancesRequest(plugins plugin.Supplier) *CommonInstancesRequest {
	value := instanceRequestPool.Get()
	if nil == value {
		req := &CommonInstancesRequest{}
		req.RouteInfo.Init(plugins)
		return req
	}
	return value.(*CommonInstancesRequest)
}

//归还到请求对象到池子
func PoolPutCommonInstancesRequest(request *CommonInstancesRequest) {
	instanceRequestPool.Put(request)
}

func PoolGetCommonServiceCallResultRequest(plugins plugin.Supplier) *CommonServiceCallResultRequest {
	value := serviceCallResultRequestPool.Get()
	if nil == value {
		req := &CommonServiceCallResultRequest{}
		return req
	}
	return value.(*CommonServiceCallResultRequest)
}

func PoolPutCommonServiceCallResultRequest(request *CommonServiceCallResultRequest) {
	serviceCallResultRequestPool.Put(request)
}

//通过池子获取请求对象
func PoolGetCommonRuleRequest() *CommonRuleRequest {
	value := ruleRequestPool.Get()
	if nil == value {
		return &CommonRuleRequest{}
	}
	return value.(*CommonRuleRequest)
}

//归还到请求对象到池子
func PoolPutCommonRuleRequest(request *CommonRuleRequest) {
	ruleRequestPool.Put(request)
}

//通过池子获取请求对象
func PoolGetCommonRateLimitRequest() *CommonRateLimitRequest {
	value := rateLimitRequestPool.Get()
	if nil == value {
		return &CommonRateLimitRequest{}
	}
	return value.(*CommonRateLimitRequest)
}

//归还到请求对象到池子
func PoolPutCommonRateLimitRequest(request *CommonRateLimitRequest) {
	rateLimitRequestPool.Put(request)
}

//通用请求对象，主要用于在消息过程减少GC
type CommonInstancesRequest struct {
	FlowID          uint64
	DstService      model.ServiceKey
	SrcService      model.ServiceKey
	Trigger         model.NotifyTrigger
	HasSrcService   bool
	DoLoadBalance   bool
	RouteInfo       servicerouter.RouteInfo
	DstInstances    model.ServiceInstances
	Revision        string
	Criteria        loadbalancer.Criteria
	FetchAll        bool
	SkipRouteFilter bool
	ControlParam    model.ControlParam
	CallResult      model.APICallResult
	response        *model.InstancesResponse
	//负载均衡算法
	LbPolicy string
}

//清理请求体
func (c *CommonInstancesRequest) clearValues(cfg config.Configuration) {
	c.FlowID = 0
	c.RouteInfo.ClearValue()
	c.DstInstances = nil
	c.Criteria.HashValue = 0
	c.Criteria.HashKey = nil
	c.Criteria.Cluster = nil
	c.Trigger.Clear()
	c.Criteria.ReplicateInfo.Count = 0
	c.Criteria.ReplicateInfo.Nodes = nil
	c.DoLoadBalance = false
	c.HasSrcService = false
	c.SkipRouteFilter = false
	c.FetchAll = false
	c.response = nil
	c.LbPolicy = ""
}

//通过获取单个请求初始化通用请求对象
func (c *CommonInstancesRequest) InitByGetOneRequest(request *model.GetOneInstanceRequest, cfg config.Configuration) {
	c.clearValues(cfg)
	c.FlowID = request.FlowID
	c.DstService.Service = request.Service
	c.DstService.Namespace = request.Namespace
	c.RouteInfo.DestService = request
	c.RouteInfo.Canary = request.Canary
	c.response = request.GetResponse()
	c.DoLoadBalance = true
	srcService := request.SourceService
	c.Trigger.EnableDstInstances = true
	c.Trigger.EnableDstRoute = true
	if nil != srcService {
		c.HasSrcService = true
		c.SrcService.Namespace = srcService.Namespace
		c.SrcService.Service = srcService.Service
		c.RouteInfo.SourceService = srcService
		if len(srcService.Namespace) > 0 && len(srcService.Service) > 0 {
			c.Trigger.EnableSrcRoute = true
		}
	}
	c.Criteria.HashKey = request.HashKey
	c.Criteria.HashValue = request.HashValue
	c.Criteria.ReplicateInfo.Count = request.ReplicateCount
	c.CallResult.APIName = model.ApiGetOneInstance
	c.CallResult.RetStatus = model.RetSuccess
	c.CallResult.RetCode = model.ErrCodeSuccess
	c.LbPolicy = request.LbPolicy
	BuildControlParam(request, cfg, &c.ControlParam)
}

//通过获取多个请求初始化通用请求对象
func (c *CommonInstancesRequest) InitByGetMultiRequest(request *model.GetInstancesRequest, cfg config.Configuration) {
	c.clearValues(cfg)
	c.FlowID = request.FlowID
	c.DstService.Service = request.Service
	c.DstService.Namespace = request.Namespace
	c.RouteInfo.DestService = request
	c.RouteInfo.Canary = request.Canary
	c.response = request.GetResponse()
	c.SkipRouteFilter = request.SkipRouteFilter
	srcService := request.SourceService
	c.Trigger.EnableDstInstances = true
	c.Trigger.EnableDstRoute = true
	if nil != srcService {
		c.HasSrcService = true
		c.SrcService.Namespace = srcService.Namespace
		c.SrcService.Service = srcService.Service
		c.RouteInfo.SourceService = srcService
		if len(srcService.Namespace) > 0 && len(srcService.Service) > 0 {
			c.Trigger.EnableSrcRoute = true
		}
	}
	c.CallResult.APIName = model.ApiGetInstances
	c.CallResult.RetStatus = model.RetSuccess
	c.CallResult.RetCode = model.ErrCodeSuccess
	BuildControlParam(request, cfg, &c.ControlParam)
}

//通过获取全部请求初始化通用请求对象
func (c *CommonInstancesRequest) InitByGetAllRequest(request *model.GetAllInstancesRequest, cfg config.Configuration) {
	c.clearValues(cfg)
	c.FlowID = request.FlowID
	c.DstService.Service = request.Service
	c.DstService.Namespace = request.Namespace
	c.RouteInfo.DestService = request
	c.response = request.GetResponse()
	c.FetchAll = true
	c.Trigger.EnableDstInstances = true
	c.CallResult.APIName = model.ApiGetAllInstances
	c.CallResult.RetStatus = model.RetSuccess
	c.CallResult.RetCode = model.ErrCodeSuccess
	BuildControlParam(request, cfg, &c.ControlParam)
}

//通过重定向服务来进行刷新
func (c *CommonInstancesRequest) RefreshByRedirect(redirectedService *model.ServiceInfo) {
	c.DstService.Namespace = redirectedService.Namespace
	c.DstService.Service = redirectedService.Service
	c.Trigger.EnableDstInstances = true
	c.Trigger.EnableDstRoute = true
	c.RouteInfo.DestRouteRule = nil
	c.DstInstances = nil
}

//构建查询实例的应答
func (c *CommonInstancesRequest) BuildInstancesResponse(flowID uint64, dstService model.ServiceKey,
	cluster *model.Cluster, instances []model.Instance, totalWeight int, revision string,
	serviceMetaData map[string]string) *model.InstancesResponse {
	return buildInstancesResponse(c.response, flowID, dstService, cluster, instances, totalWeight, revision,
		serviceMetaData)
}

//获取目标服务
func (c *CommonInstancesRequest) GetDstService() *model.ServiceKey {
	return &c.DstService
}

//获取源服务
func (c *CommonInstancesRequest) GetSrcService() *model.ServiceKey {
	return &c.SrcService
}

//获取缓存查询触发器
func (c *CommonInstancesRequest) GetNotifierTrigger() *model.NotifyTrigger {
	return &c.Trigger
}

//设置目标服务实例
func (c *CommonInstancesRequest) SetDstInstances(instances model.ServiceInstances) {
	c.DstInstances = instances
	c.Revision = instances.GetRevision()
}

//设置目标服务路由规则
func (c *CommonInstancesRequest) SetDstRoute(rule model.ServiceRule) {
	c.RouteInfo.DestRouteRule = rule
}

//设置目标服务限流规则
func (c *CommonInstancesRequest) SetDstRateLimit(rule model.ServiceRule) {
	//do nothing
}

//设置源服务路由规则
func (c *CommonInstancesRequest) SetSrcRoute(rule model.ServiceRule) {
	c.RouteInfo.SourceRouteRule = rule
}

//获取接口调用统计结果
func (c *CommonInstancesRequest) GetCallResult() *model.APICallResult {
	return &c.CallResult
}

//获取API调用控制参数
func (c *CommonInstancesRequest) GetControlParam() *model.ControlParam {
	return &c.ControlParam
}

//获取单个实例数组的持有者
type SingleInstancesOwner interface {
	//获取单个实例数组引用
	SingleInstances() []model.Instance
}

//构建查询实例的应答
func buildInstancesResponse(response *model.InstancesResponse, flowID uint64, dstService model.ServiceKey,
	cluster *model.Cluster, instances []model.Instance, totalWeight int, revision string,
	serviceMetaData map[string]string) *model.InstancesResponse {
	response.FlowID = flowID
	response.ServiceInfo.Service = dstService.Service
	response.ServiceInfo.Namespace = dstService.Namespace
	response.ServiceInfo.Metadata = serviceMetaData
	if nil != cluster {
		//对外返回的cluster，无需池化，因为可能会被别人引用
		cluster.SetReuse(false)
	}
	response.Cluster = cluster
	response.TotalWeight = totalWeight
	response.Instances = instances
	response.Revision = revision
	return response
}

//通用规则查询请求
type CommonRuleRequest struct {
	FlowID       uint64
	DstService   model.ServiceEventKey
	ControlParam model.ControlParam
	CallResult   model.APICallResult
	response     *model.ServiceRuleResponse
}

//清理请求体
func (cr *CommonRuleRequest) clearValues(cfg config.Configuration) {
	cr.FlowID = 0
	cr.response = nil
}

//通过获取路由规则请求初始化通用请求对象
func (cr *CommonRuleRequest) InitByGetRuleRequest(
	eventType model.EventType, request *model.GetServiceRuleRequest, cfg config.Configuration) {
	cr.clearValues(cfg)
	cr.FlowID = request.FlowID
	cr.CallResult.APIName = model.ApiGetRouteRule
	cr.CallResult.RetStatus = model.RetSuccess
	cr.CallResult.RetCode = model.ErrCodeSuccess
	cr.DstService.Namespace = request.Namespace
	cr.DstService.Service = request.Service
	cr.DstService.Type = eventType
	cr.response = request.GetResponse()
	BuildControlParam(request, cfg, &cr.ControlParam)
}

//构建规则查询应答
func (cr *CommonRuleRequest) BuildServiceRuleResponse(rule model.ServiceRule) *model.ServiceRuleResponse {
	resp := cr.response
	resp.Type = rule.GetType()
	resp.Value = rule.GetValue()
	resp.Revision = rule.GetRevision()
	resp.RuleCache = rule.GetRuleCache()
	resp.Service.Service = cr.DstService.Service
	resp.Service.Namespace = cr.DstService.Namespace
	resp.ValidateError = rule.GetValidateError()
	return resp
}

//获取接口调用统计结果
func (cr *CommonRuleRequest) GetCallResult() *model.APICallResult {
	return &cr.CallResult
}

//获取API调用控制参数
func (cr *CommonRuleRequest) GetControlParam() *model.ControlParam {
	return &cr.ControlParam
}

//通用限流接口的请求体
type CommonRateLimitRequest struct {
	DstService    model.ServiceKey
	Cluster       string
	Labels        map[string]string
	RateLimitRule model.ServiceRule
	Criteria      ratelimiter.InitCriteria
	Trigger       model.NotifyTrigger
	ControlParam  model.ControlParam
	CallResult    model.APICallResult
}

//清理请求体
func (cl *CommonRateLimitRequest) clearValues() {
	cl.Criteria.DstService = nil
	cl.Criteria.DstRule = nil
	cl.Trigger.Clear()
	cl.Cluster = ""
	cl.Labels = nil
}

//初始化配额获取请求
func (cl *CommonRateLimitRequest) InitByGetQuotaRequest(request *model.QuotaRequestImpl, cfg config.Configuration) {
	cl.clearValues()
	cl.DstService.Namespace = request.GetNamespace()
	cl.DstService.Service = request.GetService()
	cl.Cluster = request.GetCluster()
	cl.Labels = request.GetLabels()
	cl.Trigger.EnableDstInstances = true
	cl.Trigger.EnableDstRateLimit = true
	cl.CallResult.APIName = model.ApiGetQuota
	cl.CallResult.RetStatus = model.RetSuccess
	cl.CallResult.RetCode = model.ErrCodeSuccess
	BuildControlParam(request, cfg, &cl.ControlParam)

	//限流相关同步请求，减少重试此数和重试间隔
	if cl.ControlParam.MaxRetry > 2 {
		cl.ControlParam.MaxRetry = 2
	}
	if cl.ControlParam.RetryInterval > time.Millisecond*500 {
		cl.ControlParam.RetryInterval = time.Millisecond * 500
	}
	if cl.ControlParam.Timeout > time.Millisecond*500 {
		cl.ControlParam.Timeout = time.Millisecond * 500
	}
}

//获取目标服务
func (cl *CommonRateLimitRequest) GetDstService() *model.ServiceKey {
	return &cl.DstService
}

//获取源服务
func (cl *CommonRateLimitRequest) GetSrcService() *model.ServiceKey {
	return nil
}

//获取缓存查询触发器
func (cl *CommonRateLimitRequest) GetNotifierTrigger() *model.NotifyTrigger {
	return &cl.Trigger
}

//设置目标服务实例
func (cl *CommonRateLimitRequest) SetDstInstances(instances model.ServiceInstances) {
	cl.Criteria.DstService = instances
}

//设置目标服务路由规则
func (cl *CommonRateLimitRequest) SetDstRoute(rule model.ServiceRule) {
	//do nothing
}

//设置目标服务限流规则
func (cl *CommonRateLimitRequest) SetDstRateLimit(rule model.ServiceRule) {
	cl.RateLimitRule = rule
}

//设置源服务路由规则
func (cl *CommonRateLimitRequest) SetSrcRoute(rule model.ServiceRule) {
	//do nothing
}

//获取接口调用统计结果
func (cl *CommonRateLimitRequest) GetCallResult() *model.APICallResult {
	return &cl.CallResult
}

//获取API调用控制参数
func (cl *CommonRateLimitRequest) GetControlParam() *model.ControlParam {
	return &cl.ControlParam
}

func (cl *CommonRateLimitRequest) FormatLabelToStr(rule *namingpb.Rule) string {
	if len(cl.Labels) == 0 {
		return ""
	}
	var tmpList []string
	ruleLabels := rule.GetLabels()
	for k := range ruleLabels {
		if v, ok := cl.Labels[k]; ok {
			tmpList = append(tmpList, k+config.DefaultMapKeyValueSeparator+v)
		} else {
			log.GetBaseLogger().Errorf("CommonRateLimitRequest FormatLabelToStr not match namespace:%s "+
				"service:%s ruleId:%s notMatchKey:%s",
				rule.GetNamespace().GetValue(), rule.GetService().GetValue(), rule.GetId().GetValue(), k)
		}
	}
	sort.Strings(tmpList)
	s := strings.Join(tmpList, config.DefaultMapKVTupleSeparator)
	return s
}

type CommonServiceCallResultRequest struct {
	CallResult model.APICallResult
}

func (c *CommonServiceCallResultRequest) InitByServiceCallResult(request *model.ServiceCallResult,
	cfg config.Configuration) {
	c.CallResult.APIName = model.ApiUpdateServiceCallResult
	c.CallResult.RetStatus = model.RetSuccess
	c.CallResult.RetCode = model.ErrCodeSuccess
}
