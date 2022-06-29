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

package config

import (
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	"time"
)

//PluginConfig 插件配置对象
type PluginConfig interface {
	//cl5.*.plugin.<name>下面的所有配置
	//获取插件配置，将map[string]interface形式的配置marshal进value中
	GetPluginConfig(pluginName string) BaseConfig
	//设置插件配置，将value的内容unmarshal为map[string]interface{}形式
	SetPluginConfig(plugName string, value BaseConfig) error
	//为同一类型插件的不同实现插件创建默认配置项
	//CreateDefaultPluginCfg()
}

//GlobalConfig 全局配置对象
type GlobalConfig interface {
	PluginAwareBaseConfig
	GetSystem() SystemConfig
	//cl5.global.api前缀开头的所有配置项
	GetAPI() APIConfig
	//cl5.global.serverConnector前缀开头的所有配置项
	GetServerConnector() ServerConnectorConfig
	//cl5.global.statReporter前缀开头的所有配置项
	GetStatReporter() StatReporterConfig
}

//ConsumerConfig 调用者配置对象
type ConsumerConfig interface {
	PluginAwareBaseConfig
	//cl5.consumer.localCache前缀开头的所有配置
	GetLocalCache() LocalCacheConfig
	//cl5.consumer.serviceRouter前缀开头的所有配置
	GetServiceRouter() ServiceRouterConfig
	//cl5.consumer.loadbalancer前缀开头的所有配置
	GetLoadbalancer() LoadbalancerConfig
	//cl5.consumer.circuitbreaker前缀开头的所有配置
	GetCircuitBreaker() CircuitBreakerConfig
	//cl5.consumer.outlierDetection前缀开头的所有配置
	GetOutlierDetectionConfig() OutlierDetectionConfig
	//订阅配置
	GetSubScribe() SubscribeConfig
	//服务独立配置
	GetServiceSpecific(namespace string, service string) ServiceSpecificConfig
}

//ProviderConfig 被调端配置对象
type ProviderConfig interface {
	PluginAwareBaseConfig
	//获取限流配置
	GetRateLimit() RateLimitConfig
}

//限流相关配置
type RateLimitConfig interface {
	PluginAwareBaseConfig
	PluginConfig
	//是否启用限流能力
	IsEnable() bool
	//设置是否启用限流能力
	SetEnable(bool)
	//返回限流行为使用的插件
	//GetBehaviorPlugin(behavior RateLimitBehavior) string
	//设置限流行为使用的插件
	//SetBehaviorPlugin(behavior RateLimitBehavior, p string)
	SetMode(string)

	GetMode() model.ConfigMode

	SetRateLimitCluster(namespace string, service string)

	GetRateLimitCluster() ServerClusterConfig
	//获取最大限流窗口数量
	GetMaxWindowSize() int
	//设置最大限流窗口数量
	SetMaxWindowSize(maxSize int)
}

//系统配置信息
type SystemConfig interface {
	BaseConfig
	//global.systemConfig.mode
	//SDK运行模式，agent还是noagent
	GetMode() model.RunMode
	//设置SDK运行模式
	SetMode(model.RunMode)
	//global.systemConfig.discoverCluster
	//服务发现集群
	GetDiscoverCluster() ServerClusterConfig
	//global.systemConfig.healthCheckCluster
	//健康检查集群
	GetHealthCheckCluster() ServerClusterConfig
	//global.systemConfig.monitorCluster
	//监控上报集群
	GetMonitorCluster() ServerClusterConfig
	//global.systemConfig.rateLimitCluster
	//监控上报集群
	GetRateLimitCluster() ServerClusterConfig
	//global.systemConfig.metricCluster
	//分布式熔断、限流集群
	GetMetricCluster() ServerClusterConfig
}

//单个系统服务集群
type ServerClusterConfig interface {
	BaseConfig
	//获取命名空间
	GetNamespace() string
	//设置命名空间
	SetNamespace(string)
	//获取服务名
	GetService() string
	//设置服务名
	SetService(string)
	//系统服务的刷新间隔
	GetRefreshInterval() time.Duration
	//设置系统服务的刷新间隔
	SetRefreshInterval(time.Duration)
}

//APIConfig api相关的配置对象
type APIConfig interface {
	BaseConfig
	//cl5.global.api.timeout
	//默认调用超时时间
	GetTimeout() time.Duration
	//设置默认调用超时时间
	SetTimeout(time.Duration)
	//cl5.global.api.bindIf
	//默认客户端绑定的网卡地址
	GetBindIntf() string
	//设置默认客户端绑定的网卡地址
	SetBindIntf(string)
	//cl5.global.api.bindIP
	//默认客户端绑定的IP地址
	GetBindIP() string
	//设置默认客户端绑定的IP地址
	SetBindIP(string)
	//cl5.global.api.reportInterval
	//默认客户端定时上报周期
	GetReportInterval() time.Duration
	//设置默认客户端定时上报周期
	SetReportInterval(time.Duration)
	//cl5.global.api.maxRetryTimes
	//api调用最多重试时间
	GetMaxRetryTimes() int
	//设置api调用最多重试时间
	SetMaxRetryTimes(int)
	//cl5.global.api.retryInterval
	//api调用重试时间
	GetRetryInterval() time.Duration
	//设置api调用重试时间
	SetRetryInterval(time.Duration)
}

//统计上报配置
type StatReporterConfig interface {
	PluginAwareBaseConfig
	PluginConfig
	//是否启用上报
	IsEnable() bool
	//设置是否启用上报
	SetEnable(bool)
	//统计上报器插件链
	GetChain() []string
	//设置统计上报器插件链
	SetChain([]string)
}

//ServerConnectorConfig 与名字服务服务端的连接配置
type ServerConnectorConfig interface {
	PluginAwareBaseConfig
	PluginConfig
	//global.serverConnector.addresses
	//远端server地址，格式为<host>:<port>
	GetAddresses() []string
	//设置远端server地址，格式为<host>:<port>
	SetAddresses([]string)
	//global.serverConnector.protocol
	//与server对接的协议
	GetProtocol() string
	//设置与server对接的协议
	SetProtocol(string)
	//global.serverConnector.connectTimeout
	//与server的连接超时时间
	GetConnectTimeout() time.Duration
	//设置与server的连接超时时间
	SetConnectTimeout(time.Duration)
	//GetMessageTimeout global.registry.messageTimeout
	//远程请求超时时间
	GetMessageTimeout() time.Duration
	//设置远程请求超时时间
	SetMessageTimeout(time.Duration)
	//global.serverConnector.clientRequestQueueSize
	//新请求的队列BUFFER容量
	GetRequestQueueSize() int32
	//设置新请求的队列BUFFER容量
	SetRequestQueueSize(int32)
	//global.serverConnector.serverSwitchInterval
	// server的切换时延
	GetServerSwitchInterval() time.Duration
	// 设置server的切换时延
	SetServerSwitchInterval(time.Duration)
	//global.serverConnector.connectionIdleTimeout
	//连接会被释放的空闲的时长
	GetConnectionIdleTimeout() time.Duration
	//设置连接会被释放的空闲的时长
	SetConnectionIdleTimeout(time.Duration)
	//埋点集群类型
	GetJoinPoint() string
	//设置埋点集群类型
	SetJoinPoint(string)
}

//LocalCacheConfig 本地缓存相关配置项
type LocalCacheConfig interface {
	PluginAwareBaseConfig
	PluginConfig
	//cl5.consumer.localCache.service.expireTime,
	//服务的超时淘汰时间
	GetServiceExpireTime() time.Duration
	//设置服务的超时淘汰时间
	SetServiceExpireTime(time.Duration)
	//cl5.consumer.localCache.service.refreshInterval
	//服务的定期刷新时间
	GetServiceRefreshInterval() time.Duration
	//设置服务的定期刷新时间
	SetServiceRefreshInterval(time.Duration)
	//cl5.consumer.localCache.persistDir
	//本地缓存持久化路径
	GetPersistDir() string
	//设置本地缓存持久化路径
	SetPersistDir(string)
	//cl5.consumer.localCache.type
	//本地缓存类型，默认default，可修改成具体的缓存插件名
	GetType() string
	//设置本地缓存类型
	SetType(string)
	//consumer.localCache.persistMaxWriteRetry
	//缓存最大写重试次数
	GetPersistMaxWriteRetry() int
	//设置缓存最大写重试次数
	SetPersistMaxWriteRetry(int)
	//consumer.localCache.persistMaxReadRetry
	//缓存最大读重试次数
	GetPersistMaxReadRetry() int
	//设置缓存最大读重试次数
	SetPersistMaxReadRetry(int)
	//consumer.localCache.persistRetryInterval
	//缓存持久化重试间隔
	GetPersistRetryInterval() time.Duration
	//设置缓存持久化重试间隔
	SetPersistRetryInterval(time.Duration)
}

//就近路由配置
type NearbyConfig interface {
	BaseConfig
	//设置匹配级别，consumer.serviceRouter.plugin.nearbyBasedRouter.matchLevel
	SetMatchLevel(level string)
	//获取匹配级别，consumer.serviceRouter.plugin.nearbyBasedRouter.matchLevel
	GetMatchLevel() string
	//设置可以降级的最低匹配级别,consumer.serviceRouter.plugin.nearbyBasedRouter.lowestMatchLevel
	SetLowestMatchLevel(level string)
	//获取可以降级的最低匹配级别,consumer.serviceRouter.plugin.nearbyBasedRouter.lowestMatchLevel
	GetLowestMatchLevel() string
	//设置是否进行严格就近,consumer.serviceRouter.plugin.nearbyBasedRouter.strictNearby
	SetStrictNearby(s bool)
	//获取是否进行严格就近,consumer.serviceRouter.plugin.nearbyBasedRouter.strictNearby
	IsStrictNearby() bool
	//是否开启根据不健康实例比例进行降级就近匹配,
	//consumer.serviceRouter.plugin.nearbyBasedRouter.enableDegradeByUnhealthyPercent
	IsEnableDegradeByUnhealthyPercent() bool
	//设置是否开启根据不健康实例比例进行降级就近匹配
	//consumer.serviceRouter.plugin.nearbyBasedRouter.enableDegradeByUnhealthyPercent
	SetEnableDegradeByUnhealthyPercent(e bool)
	//获取触发降级匹配的不健康实例比例,
	//consumer.serviceRouter.plugin.nearbyBasedRouter.unhealthyPercentToDegrade
	GetUnhealthyPercentToDegrade() int
	//设置触发降级匹配的不健康实例比例,consumer.serviceRouter.plugin.nearbyBasedRouter.unhealthyPercentToDegrade
	SetUnhealthyPercentToDegrade(u int)
}

//ServiceRouterConfig 服务路由相关配置项
type ServiceRouterConfig interface {
	PluginAwareBaseConfig
	PluginConfig
	//cl5.consumer.serviceRouter.chain
	//路由责任链配置
	GetChain() []string
	//设置路由责任链配置
	SetChain([]string)
	//获取PercentOfMinInstances参数
	GetPercentOfMinInstances() float64
	//设置PercentOfMinInstances参数
	SetPercentOfMinInstances(float64)
	//是否启用全死全活机制
	IsEnableRecoverAll() bool
	//设置启用全死全活机制
	SetEnableRecoverAll(bool)
	//获取就近路由配置
	GetNearbyConfig() NearbyConfig
}

//LoadbalancerConfig 负载均衡相关配置项
type LoadbalancerConfig interface {
	PluginAwareBaseConfig
	PluginConfig
	//负载均衡类型
	GetType() string
	//设置负载均衡类型
	SetType(string)
}

//CircuitBreakerConfig 熔断相关的配置项
type CircuitBreakerConfig interface {
	PluginAwareBaseConfig
	//是否启用熔断
	IsEnable() bool
	//设置是否启用熔断
	SetEnable(bool)
	//熔断器插件链
	GetChain() []string
	//设置熔断器插件链
	SetChain([]string)
	//熔断器定时检测时间
	GetCheckPeriod() time.Duration
	//设置熔断器定时检测时间
	SetCheckPeriod(time.Duration)
	//获取熔断周期
	GetSleepWindow() time.Duration
	//设置熔断周期
	SetSleepWindow(interval time.Duration)
	//获取半开状态后最多分配多少个探测请求
	GetRequestCountAfterHalfOpen() int
	//设置半开状态后最多分配多少个探测请求
	SetRequestCountAfterHalfOpen(count int)
	//获取半开状态后多少个成功请求则恢复
	GetSuccessCountAfterHalfOpen() int
	//设置半开状态后多少个成功请求则恢复
	SetSuccessCountAfterHalfOpen(count int)
	//获取半开后的恢复周期，按周期来进行半开放量的统计
	GetRecoverWindow() time.Duration
	//设置半开后的恢复周期，按周期来进行半开放量的统计
	SetRecoverWindow(value time.Duration)
	//半开后请求数统计滑桶数量
	GetRecoverNumBuckets() int
	//设置半开后请求数统计滑桶数量
	SetRecoverNumBuckets(value int)
	//连续错误数熔断配置
	GetErrorCountConfig() ErrorCountConfig
	//错误率熔断配置
	GetErrorRateConfig() ErrorRateConfig
}

//Configuration 全量配置对象
type Configuration interface {
	PluginAwareBaseConfig
	//cl5.global前缀开头的所有配置项
	GetGlobal() GlobalConfig
	//cl5.consumer前缀开头的所有配置项
	GetConsumer() ConsumerConfig
	//cl5.provider前缀开头的所有配置项
	GetProvider() ProviderConfig
}

//主动健康检查配置
type OutlierDetectionConfig interface {
	PluginAwareBaseConfig
	PluginConfig
	//定时探测时间
	GetCheckPeriod() time.Duration
	//设置定时探测时间
	SetCheckPeriod(time.Duration)
	//是否启用健康检查
	IsEnable() bool
	//设置是否启用健康检查
	SetEnable(bool)
	//健康检查插件链
	GetChain() []string
	//设置健康检查插件链
	SetChain([]string)
}

type SubscribeConfig interface {
	PluginAwareBaseConfig
	PluginConfig
	//获取插件
	GetType() string
	//设置插件
	SetType(string)
}

type ServiceSpecificConfig interface {
	BaseConfig

	GetServiceCircuitBreaker() CircuitBreakerConfig

	GetServiceRouter() ServiceRouterConfig
}
