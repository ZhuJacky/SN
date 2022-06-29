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
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"git.code.oa.com/polaris/polaris-go/pkg/model"
	"gopkg.in/yaml.v2"
)

//ConfigurationImpl cl5全局配置
type ConfigurationImpl struct {
	Global   *GlobalConfigImpl   `yaml:"global" json:"global"`
	Consumer *ConsumerConfigImpl `yaml:"consumer" json:"consumer"`
	Provider *ProviderConfigImpl `yaml:"provider" json:"provider"`
}

//GetGlobal cl5.global前缀开头的所有配置项
func (c *ConfigurationImpl) GetGlobal() GlobalConfig {
	return c.Global
}

//GetConsumer cl5.consumer前缀开头的所有配置项
func (c *ConfigurationImpl) GetConsumer() ConsumerConfig {
	return c.Consumer
}

//GetConsumer consumer前缀开头的所有配置项
func (c *ConfigurationImpl) GetProvider() ProviderConfig {
	return c.Provider
}

//获取该域下所有插件的名字
func (c *ConfigurationImpl) GetPluginNames() model.HashSet {
	nameMap := model.HashSet{}
	memberArr := []PluginAwareBaseConfig{c.Global, c.Consumer, c.Provider}
	for _, member := range memberArr {
		for k := range member.GetPluginNames() {
			nameMap.Add(k)
		}
	}
	return nameMap
}

// 全局配置
type GlobalConfigImpl struct {
	System          *SystemConfigImpl          `yaml:"system" json:"system"`
	API             *APIConfigImpl             `yaml:"api" json:"api"`
	ServerConnector *ServerConnectorConfigImpl `yaml:"serverConnector" json:"serverConnector"`
	StatReporter    *StatReporterConfigImpl    `yaml:"statReporter" json:"statReporter"`
}

//获取该域下所有插件的名字
func (g *GlobalConfigImpl) GetPluginNames() model.HashSet {
	nameMap := model.HashSet{}
	memberArr := []PluginAwareBaseConfig{g.ServerConnector, g.StatReporter}
	for _, member := range memberArr {
		for k := range member.GetPluginNames() {
			nameMap.Add(k)
		}
	}
	return nameMap
}

//获取系统配置
func (g *GlobalConfigImpl) GetSystem() SystemConfig {
	return g.System
}

//GetAPI global.api前缀开头的所有配置项
func (g *GlobalConfigImpl) GetAPI() APIConfig {
	return g.API
}

//GetServerConnector global.serverConnector前缀开头的所有配置项
func (g *GlobalConfigImpl) GetServerConnector() ServerConnectorConfig {
	return g.ServerConnector
}

//cl5.global.statReporter前缀开头的所有配置项
func (g *GlobalConfigImpl) GetStatReporter() StatReporterConfig {
	return g.StatReporter
}

// 消费者配置
type ConsumerConfigImpl struct {
	LocalCache       *LocalCacheConfigImpl       `yaml:"localCache" json:"localCache"`
	ServiceRouter    *ServiceRouterConfigImpl    `yaml:"serviceRouter" json:"serviceRouter"`
	Loadbalancer     *LoadBalancerConfigImpl     `yaml:"loadbalancer" json:"loadbalancer"`
	CircuitBreaker   *CircuitBreakerConfigImpl   `yaml:"circuitBreaker" json:"circuitBreaker"`
	OutlierDetection *OutlierDetectionConfigImpl `yaml:"outlierDetection" json:"outlierDetection"`
	Subscribe        *SubscribeImpl              `yaml:"subscribe" json:"subscribe"`
	ServicesSpecific []*ServiceSpecific          `yaml:"servicesSpecific" json:"servicesSpecific"`
}

//GetLocalCache consumer.localCache前缀开头的所有配置
func (c *ConsumerConfigImpl) GetLocalCache() LocalCacheConfig {
	return c.LocalCache
}

//GetServiceRouter consumer.serviceRouter前缀开头的所有配置
func (c *ConsumerConfigImpl) GetServiceRouter() ServiceRouterConfig {
	return c.ServiceRouter
}

//GetLoadbalancer consumer.loadbalancer前缀开头的所有配置
func (c *ConsumerConfigImpl) GetLoadbalancer() LoadbalancerConfig {
	return c.Loadbalancer
}

//GetLoadbalancer consumer.circuitbreaker前缀开头的所有配置
func (c *ConsumerConfigImpl) GetCircuitBreaker() CircuitBreakerConfig {
	return c.CircuitBreaker
}

//cl5.consumer.outlierDetection前缀开头的所有配置
func (c *ConsumerConfigImpl) GetOutlierDetectionConfig() OutlierDetectionConfig {
	return c.OutlierDetection
}

//订阅配置
func (c *ConsumerConfigImpl) GetSubScribe() SubscribeConfig {
	return c.Subscribe
}

//获取该域下所有插件的名字
func (c *ConsumerConfigImpl) GetPluginNames() model.HashSet {
	nameMap := model.HashSet{}
	memberArr := []PluginAwareBaseConfig{c.LocalCache, c.ServiceRouter, c.Loadbalancer,
		c.CircuitBreaker, c.OutlierDetection}
	for _, member := range memberArr {
		for k := range member.GetPluginNames() {
			nameMap.Add(k)
		}
	}
	return nameMap
}

// 设置 discover 服务配置
func (c *ConsumerConfigImpl) setPolarisDiscoverSpecific() {
	v := c.GetServiceSpecific(ServerNamespace, ServerDiscoverService)
	if v != nil {
		return
	}
	sConf := &ServiceSpecific{
		Namespace: ServerNamespace,
		Service:   ServerDiscoverService,
	}
	sConf.Init()
	sConf.SetDefault()
	sConf.CircuitBreaker.GetErrorCountConfig().SetContinuousErrorThreshold(1)
	sConf.ServiceRouter.GetNearbyConfig().SetMatchLevel(RegionLevel)
	c.ServicesSpecific = append(c.ServicesSpecific, sConf)
}

//服务独立配置
func (c *ConsumerConfigImpl) GetServiceSpecific(namespace string, service string) ServiceSpecificConfig {
	for _, v := range c.ServicesSpecific {
		if v.Namespace == namespace && v.Service == service {
			return v
		}
	}
	return nil
}

//系统配置
type SystemConfigImpl struct {
	//SDK运行模式
	Mode model.RunMode `yaml:"mode" json:"mode"`
	//服务发现集群
	DiscoverCluster *ServerClusterConfigImpl `yaml:"discoverCluster" json:"discoverCluster"`
	//健康检查集群
	HealthCheckCluster *ServerClusterConfigImpl `yaml:"healthCheckCluster" json:"healthCheckCluster"`
	//监控上报集群
	MonitorCluster *ServerClusterConfigImpl `yaml:"monitorCluster" json:"monitorCluster"`
	//服务限流集群
	RateLimitCluster *ServerClusterConfigImpl `yaml:"rateLimitCluster" json:"rateLimitCluster"`
	//metric集群（用于限流、熔断）
	MetricCluster *ServerClusterConfigImpl `yaml:"metricCluster" json:"metricCluster"`
}

//SDK运行模式，agent还是noagent
func (s *SystemConfigImpl) GetMode() model.RunMode {
	return s.Mode
}

//设置SDK运行模式
func (s *SystemConfigImpl) SetMode(mode model.RunMode) {
	s.Mode = mode
}

//服务发现集群
func (s *SystemConfigImpl) GetDiscoverCluster() ServerClusterConfig {
	return s.DiscoverCluster
}

//健康检查集群
func (s *SystemConfigImpl) GetHealthCheckCluster() ServerClusterConfig {
	return s.HealthCheckCluster
}

//监控上报集群
func (s *SystemConfigImpl) GetMonitorCluster() ServerClusterConfig {
	return s.MonitorCluster
}

//服务限流集群
func (s *SystemConfigImpl) GetRateLimitCluster() ServerClusterConfig {
	return s.RateLimitCluster
}

func (s *SystemConfigImpl) GetMetricCluster() ServerClusterConfig {
	return s.MetricCluster
}

//单个服务集群配置
type ServerClusterConfigImpl struct {
	Namespace       string         `yaml:"namespace" json:"namespace"`
	Service         string         `yaml:"service" json:"service"`
	RefreshInterval *time.Duration `yaml:"refreshInterval" json:"refreshInterval"`
}

//获取命名空间
func (s *ServerClusterConfigImpl) GetNamespace() string {
	return s.Namespace
}

//设置命名空间
func (s *ServerClusterConfigImpl) SetNamespace(namespace string) {
	s.Namespace = namespace
}

//获取服务名
func (s *ServerClusterConfigImpl) GetService() string {
	return s.Service
}

//设置服务名
func (s *ServerClusterConfigImpl) SetService(service string) {
	s.Service = service
}

//获取系统服务刷新间隔
func (s *ServerClusterConfigImpl) GetRefreshInterval() time.Duration {
	return *s.RefreshInterval
}

//获取系统服务刷新间隔
func (s *ServerClusterConfigImpl) SetRefreshInterval(interval time.Duration) {
	s.RefreshInterval = &interval
}

//通过服务信息创建服务集群配置
func NewServerClusterConfig(svcKey model.ServiceKey) *ServerClusterConfigImpl {
	return &ServerClusterConfigImpl{
		Namespace: svcKey.Namespace,
		Service:   svcKey.Service,
	}
}

//服务集群信息转换为服务对象
func ServiceClusterToServiceKey(config ServerClusterConfig) model.ServiceKey {
	return model.ServiceKey{
		Namespace: config.GetNamespace(),
		Service:   config.GetService(),
	}
}

// API访问相关的配置
type APIConfigImpl struct {
	Timeout        *time.Duration `yaml:"timeout" json:"timeout"`
	BindIntf       string         `yaml:"bindIf" json:"bindIf"`
	BindIP         string         `yaml:"bindIP" json:"bindIP"`
	BindIPValue    string         `yaml:"-" json:"-"`
	ReportInterval *time.Duration `yaml:"reportInterval" json:"reportInterval"`
	MaxRetryTimes  int            `yaml:"maxRetryTimes" json:"maxRetryTimes"`
	RetryInterval  *time.Duration `yaml:"retryInterval" json:"retryInterval"`
}

//GetTimeout global.api.timeout
// 默认调用超时时间
func (a *APIConfigImpl) GetTimeout() time.Duration {
	return *a.Timeout
}

//设置默认超时时间
func (a *APIConfigImpl) SetTimeout(timeout time.Duration) {
	a.Timeout = &timeout
}

//GetBindIntf global.api.bindIf
// 默认客户端绑定的网卡地址
func (a *APIConfigImpl) GetBindIntf() string {
	return a.BindIntf
}

//设置默认客户端绑定的网卡地址
func (a *APIConfigImpl) SetBindIntf(bindIntf string) {
	a.BindIntf = bindIntf
}

//GetBindIntf global.api.bindIP
// 默认客户端绑定的网卡地址
func (a *APIConfigImpl) GetBindIP() string {
	return a.BindIPValue
}

//设置默认客户端绑定的网卡地址
func (a *APIConfigImpl) SetBindIP(bindIPValue string) {
	a.BindIPValue = bindIPValue
}

//GetReportInterval global.api.reportInterval
// 默认客户端上报周期
func (a *APIConfigImpl) GetReportInterval() time.Duration {
	return *a.ReportInterval
}

// 设置默认客户端上报周期
func (a *APIConfigImpl) SetReportInterval(interval time.Duration) {
	a.ReportInterval = &interval
}

//最大重试次数
func (a *APIConfigImpl) GetMaxRetryTimes() int {
	return a.MaxRetryTimes
}

//最大重试次数
func (a *APIConfigImpl) SetMaxRetryTimes(maxRetryTimes int) {
	a.MaxRetryTimes = maxRetryTimes
}

//重试周期
func (a *APIConfigImpl) GetRetryInterval() time.Duration {
	return *a.RetryInterval
}

//重试周期
func (a *APIConfigImpl) SetRetryInterval(interval time.Duration) {
	a.RetryInterval = &interval
}

//创建默认配置对象
func NewDefaultConfiguration(serverAddrs []string) *ConfigurationImpl {
	cfg := &ConfigurationImpl{}
	cfg.Init()
	cfg.SetDefault()
	cfg.GetGlobal().GetServerConnector().(*ServerConnectorConfigImpl).Addresses = serverAddrs
	return cfg
}

//根据埋点ip的端口构造埋点Server列表
func getDefaultBuiltinServers(joinPoint string) []string {
	if builtInServers, ok := joinPointsBuiltInServers[joinPoint]; ok {
		res := make([]string, len(builtInServers))
		for i := 0; i < len(res); i++ {
			res[i] = fmt.Sprintf("%s:%d", model.ToNetIP(builtInServers[i]), defaultBuiltinServerPort)
		}
		return res
	} else {
		panic(fmt.Sprintf("%s not find in joinPointsBuiltInServers", joinPoint))
	}
}

//获取可以从获取的容器
func GetContainerNameEnvList() []string {
	res := make([]string, len(containerNameEnvs))
	for i, c := range containerNameEnvs {
		res[i] = c
	}
	return res
}

//创建带有默认埋点server域名的默认配置
func NewDefaultConfigurationWithDomain() *ConfigurationImpl {
	return NewDefaultConfiguration(getDefaultBuiltinServers(JoinPointMainland))
}

//LoadConfigurationByFile 通过文件加载配置项
func LoadConfigurationByFile(path string) (*ConfigurationImpl, error) {
	if !model.IsFile(path) {
		return nil, model.NewSDKError(model.ErrCodeAPIInvalidArgument, nil, "invalid context file %s", path)
	}
	buff, err := ioutil.ReadFile(path)
	if nil != err {
		return nil, model.NewSDKError(model.ErrCodeAPIInvalidArgument, err, "fail to read context file %s", path)
	}
	return LoadConfiguration(buff)
}

//通过默认配置文件加载配置项
func LoadConfigurationByDefaultFile() (*ConfigurationImpl, error) {
	path := model.ReplaceHomeVar(DefaultConfigFile)
	return LoadConfigurationByFile(path)
}

//LoadConfiguration 加载配置项
func LoadConfiguration(buf []byte) (*ConfigurationImpl, error) {
	var err error
	cfg := &ConfigurationImpl{}
	cfg.Init()
	decoder := yaml.NewDecoder(bytes.NewBuffer(buf))
	if err = decoder.Decode(cfg); nil != err {
		return nil, model.NewSDKError(model.ErrCodeAPIInvalidConfig, err,
			"fail to decode config string")
	}
	cfg.SetDefault()
	if err = cfg.Verify(); nil != err {
		return nil, model.NewSDKError(model.ErrCodeAPIInvalidConfig, err,
			"fail to verify config string")
	}

	return cfg, nil
}
