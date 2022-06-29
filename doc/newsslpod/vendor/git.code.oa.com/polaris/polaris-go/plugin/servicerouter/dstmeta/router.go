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

package dstmeta

import (
	"fmt"
	"git.code.oa.com/polaris/polaris-go/pkg/config"
	"git.code.oa.com/polaris/polaris-go/pkg/log"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/common"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/servicerouter"
)

//基于目标服务元数据的服务路由插件
type InstancesFilter struct {
	*plugin.PluginBase
	percentOfMinInstances float64
	valueCtx              model.ValueContext
	recoverAll            bool
}

//Type 插件类型
func (g *InstancesFilter) Type() common.Type {
	return common.TypeServiceRouter
}

//Name 插件名，一个类型下插件名唯一
func (g *InstancesFilter) Name() string {
	return config.DefaultServiceRouterDstMeta
}

//Init 初始化插件
func (g *InstancesFilter) Init(ctx *plugin.InitContext) error {
	// 获取最小返回实例比例
	g.PluginBase = plugin.NewPluginBase(ctx)
	g.percentOfMinInstances = ctx.Config.GetConsumer().GetServiceRouter().GetPercentOfMinInstances()
	g.recoverAll = ctx.Config.GetConsumer().GetServiceRouter().IsEnableRecoverAll()
	g.valueCtx = ctx.ValueCtx
	return nil
}

//Destroy 销毁插件，可用于释放资源
func (g *InstancesFilter) Destroy() error {
	return nil
}

//GetFilteredInstances 插件模式进行服务实例过滤，并返回过滤后的实例列表
func (g *InstancesFilter) GetFilteredInstances(routeInfo *servicerouter.RouteInfo,
	clusters model.ServiceClusters, withinCluster *model.Cluster) (*servicerouter.RouteResult, error) {
	targetCluster := model.NewCluster(clusters, withinCluster)
	dstMetadata := routeInfo.DestService.GetMetadata()
	if len(dstMetadata) > 0 {
		for metaKey, metaValue := range dstMetadata {
			targetCluster.AddMetadata(metaKey, metaValue)
		}
		targetCluster.ReloadComposeMetaValue()
		instSet := targetCluster.GetClusterValue().GetInstancesSet(true, true)
		if instSet.Count() == 0 {
			errorText := fmt.Sprintf(
				"dstmeta not match, dstService %s(namespace %s), metadata is %v",
				routeInfo.DestService.GetService(), routeInfo.DestService.GetNamespace(), dstMetadata)
			log.GetBaseLogger().Errorf(errorText)
			return nil, model.NewSDKError(model.ErrCodeDstMetaMismatch, nil, errorText)
		}
	}
	result := servicerouter.PoolGetRouteResult(g.valueCtx)
	result.OutputCluster = targetCluster
	return result, nil
}

//init 注册插件
func init() {
	plugin.RegisterPlugin(&InstancesFilter{})
}

//是否需要启动规则路由
func (g *InstancesFilter) Enable(routeInfo *servicerouter.RouteInfo) bool {
	if len(routeInfo.DestService.GetMetadata()) == 0 {
		return false
	}
	return true
}
