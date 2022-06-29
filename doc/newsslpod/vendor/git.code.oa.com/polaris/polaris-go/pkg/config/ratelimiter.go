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
	"errors"
	"fmt"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/common"
)

//限流配置对象
type RateLimitConfigImpl struct {
	//是否启动限流
	Enable *bool `yaml:"enable" json:"enable"`
	//各个限流插件的配置
	Plugin PluginConfigs `yaml:"plugin" json:"plugin"`
	// mode  0: local  1: global
	Mode string `yaml:"mode" json:"mode"`
	// rateLimitCluster
	RateLimitCluster *ServerClusterConfigImpl `yaml:"rateLimitCluster" json:"rateLimitCluster"`
	//最大限流窗口数量
	MaxWindowSize int `yaml:"maxWindowSize" json:"maxWindowSize"`
}

//是否启用限流能力
func (r *RateLimitConfigImpl) IsEnable() bool {
	return *r.Enable
}

//设置是否启用限流能力
func (r *RateLimitConfigImpl) SetEnable(value bool) {
	r.Enable = &value
}

//校验配置参数
func (r *RateLimitConfigImpl) Verify() error {
	if nil == r {
		return errors.New("OutlierDetectionConfig is nil")
	}
	if nil == r.Enable {
		return fmt.Errorf("provider.rateLimit.enable must not be nil")
	}
	if r.RateLimitCluster != nil {
		if r.RateLimitCluster.GetNamespace() == ServerNamespace &&
			r.RateLimitCluster.GetService() == ServerMetricService {
			return errors.New("RateLimitCluster can not set to polaris.metric")
		}
	}
	return r.Plugin.Verify()
}

//获取插件配置
func (r *RateLimitConfigImpl) GetPluginConfig(pluginName string) BaseConfig {
	cfgValue, ok := r.Plugin[pluginName]
	if !ok {
		return nil
	}
	return cfgValue.(BaseConfig)
}

//设置默认参数
func (r *RateLimitConfigImpl) SetDefault() {
	if nil == r.Enable {
		r.Enable = &DefaultRateLimitEnable
	}
	if 0 == r.MaxWindowSize {
		r.MaxWindowSize = MaxRateLimitWindowSize
	}
	r.Plugin.SetDefault(common.TypeRateLimiter)
	r.RateLimitCluster.SetDefault()
}

//设置插件配置
func (r *RateLimitConfigImpl) SetPluginConfig(pluginName string, value BaseConfig) error {
	return r.Plugin.SetPluginConfig(common.TypeRateLimiter, pluginName, value)
}

func (r *RateLimitConfigImpl) SetMode(mode string) {
	r.Mode = mode
}

func (r *RateLimitConfigImpl) GetMode() model.ConfigMode {
	if r.Mode == model.RateLimitLocal {
		return model.ConfigQuotaLocalMode
	} else if r.Mode == model.RateLimitGlobal {
		return model.ConfigQuotaGlobalMode
	} else {
		return model.ConfigQuotaLocalMode
	}
}

func (r *RateLimitConfigImpl) SetRateLimitCluster(namespace string, service string) {
	if r.RateLimitCluster == nil {
		r.RateLimitCluster = &ServerClusterConfigImpl{}
	}
	r.RateLimitCluster.SetNamespace(namespace)
	r.RateLimitCluster.SetService(service)
}

func (r *RateLimitConfigImpl) GetRateLimitCluster() ServerClusterConfig {
	return r.RateLimitCluster
}

//配置初始化
func (r *RateLimitConfigImpl) Init() {
	r.Plugin = PluginConfigs{}
	r.Plugin.Init(common.TypeRateLimiter)
	r.Mode = string(model.ConfigQuotaGlobalMode)
	r.RateLimitCluster = &ServerClusterConfigImpl{
		Namespace: "",
		Service:   "",
	}
}

//获取该域下所有插件的名字
func (r *RateLimitConfigImpl) GetPluginNames() model.HashSet {
	names := model.HashSet{}
	names.Add(DefaultRejectRateLimiter)
	names.Add(DefaultWarmUpRateLimiter)
	names.Add(DefaultUniformRateLimiter)
	names.Add(DefaultWarmUpWaitLimiter)
	return names
}

//GetMaxWindowSize
func (r *RateLimitConfigImpl) GetMaxWindowSize() int {
	return r.MaxWindowSize
}

//SetMaxWindowSize
func (r *RateLimitConfigImpl) SetMaxWindowSize(maxSize int) {
	r.MaxWindowSize = maxSize
}

//
////为各个插件创建空白配置
//func (r *RateLimitConfigImpl) CreateDefaultPluginCfg() {
//	if r.Plugin == nil {
//		r.Plugin = make(map[string]map[string]interface{})
//	}
//}
