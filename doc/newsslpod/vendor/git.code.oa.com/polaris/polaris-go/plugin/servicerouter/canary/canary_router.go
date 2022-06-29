package canary

import (
	"errors"
	"git.code.oa.com/polaris/polaris-go/pkg/config"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/common"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/servicerouter"
	"github.com/modern-go/reflect2"
)

const ()

//CanaryRouterFilter 金丝雀过滤器
type CanaryRouterFilter struct {
	*plugin.PluginBase
	percentOfMinInstances float64
	valueCtx              model.ValueContext
	recoverAll            bool
}

//Type 插件类型
func (g *CanaryRouterFilter) Type() common.Type {
	return common.TypeServiceRouter
}

//Name 插件名，一个类型下插件名唯一
func (g *CanaryRouterFilter) Name() string {
	return config.DefaultServiceRouterCanary
}

//Init 初始化插件
func (g *CanaryRouterFilter) Init(ctx *plugin.InitContext) error {
	// 获取最小返回实例比例
	g.PluginBase = plugin.NewPluginBase(ctx)
	g.percentOfMinInstances = ctx.Config.GetConsumer().GetServiceRouter().GetPercentOfMinInstances()
	g.recoverAll = ctx.Config.GetConsumer().GetServiceRouter().IsEnableRecoverAll()
	g.valueCtx = ctx.ValueCtx
	return nil
}

//Destroy 销毁插件，可用于释放资源
func (g *CanaryRouterFilter) Destroy() error {
	return nil
}

//CanaryRouterFilter 插件模式进行服务实例过滤，并返回过滤后的实例列表
func (g *CanaryRouterFilter) GetFilteredInstances(routeInfo *servicerouter.RouteInfo,
	clusters model.ServiceClusters, withinCluster *model.Cluster) (*servicerouter.RouteResult, error) {

	enableCanary := clusters.IsCanaryEnabled()
	if !enableCanary {
		result := servicerouter.PoolGetRouteResult(g.valueCtx)
		cls := model.NewCluster(clusters, withinCluster)
		result.OutputCluster = cls
		return result, nil
	}
	canary := routeInfo.Canary
	var result *servicerouter.RouteResult
	var err error
	if canary != "" {
		result, err = g.canaryFilter(canary, clusters, withinCluster)
	} else {
		result, err = g.noCanaryFilter(clusters, withinCluster)
	}
	if err != nil {
		//返回给外层，靠filterOnly兜底
		if result == nil || reflect2.IsNil(result) {
			result = servicerouter.PoolGetRouteResult(g.valueCtx)
		}
		cls := model.NewCluster(clusters, withinCluster)
		result.OutputCluster = cls
		result.OutputCluster.HasLimitedInstances = true
		return result, nil
	} else {
		routeInfo.SetIgnoreFilterOnlyOnEndChain(true)
		return result, nil
	}
}

//带金丝雀标签的处理过滤
func (g *CanaryRouterFilter) canaryFilter(canaryValue string,
	clusters model.ServiceClusters, withinCluster *model.Cluster) (*servicerouter.RouteResult, error) {
	availableCluster, err := g.canaryAvailableFilter(canaryValue, clusters, withinCluster)
	if err == nil && availableCluster != nil {
		result := servicerouter.PoolGetRouteResult(g.valueCtx)
		result.OutputCluster = availableCluster
		return result, nil
	}
	limitedCluster, err := g.canaryLimitedFilter(canaryValue, clusters, withinCluster)
	if err == nil && limitedCluster != nil {
		result := servicerouter.PoolGetRouteResult(g.valueCtx)
		result.OutputCluster = limitedCluster
		return result, nil
	}
	return nil, errors.New("no instances after canaryFilter")
}

//带金丝雀过滤可用实例
func (g *CanaryRouterFilter) canaryAvailableFilter(canaryValue string, clusters model.ServiceClusters,
	withinCluster *model.Cluster) (*model.Cluster, error) {
	targetCluster := model.NewCluster(clusters, withinCluster)
	targetCluster.AddMetadata(model.CanaryMetaKey, canaryValue)
	targetCluster.ReloadComposeMetaValue()
	// 返回带canary的可用实例
	instSet := targetCluster.GetClusterValue().GetInstancesSet(false, true)
	if instSet.Count() > 0 {
		return targetCluster, nil
	}
	defer targetCluster.PoolPut()
	// 返回不带canary的可用实例
	notContainMetaKeyCluster := model.NewCluster(clusters, withinCluster)
	notContainMetaKeyCluster.AddMetadata(model.CanaryMetaKey, "")
	notContainMetaKeyCluster.ReloadComposeMetaValue()
	notContainMetaKeyInstSet := notContainMetaKeyCluster.GetNotContainMetaKeyClusterValue().GetInstancesSet(false, true)
	if notContainMetaKeyInstSet.Count() > 0 {
		return notContainMetaKeyCluster, nil
	}
	defer notContainMetaKeyCluster.PoolPut()
	// 找 不匹配特定金丝雀 的可用实例
	notMatchMetaKeyCluster := model.NewCluster(clusters, withinCluster)
	notMatchMetaKeyCluster.AddMetadata(model.CanaryMetaKey, canaryValue)
	notMatchMetaKeyCluster.ReloadComposeMetaValue()
	notMatchMetaKeyInstSet := notMatchMetaKeyCluster.GetContainNotMatchMetaKeyClusterValue().GetInstancesSet(false, true)
	if notMatchMetaKeyInstSet.Count() > 0 {
		return notMatchMetaKeyCluster, nil
	}
	defer notMatchMetaKeyCluster.PoolPut()
	return nil, errors.New("no available instances")
}

//带金丝雀过滤limited实例
func (g *CanaryRouterFilter) canaryLimitedFilter(canaryValue string, clusters model.ServiceClusters,
	withinCluster *model.Cluster) (*model.Cluster, error) {
	targetCluster := model.NewCluster(clusters, withinCluster)
	targetCluster.AddMetadata(model.CanaryMetaKey, canaryValue)
	targetCluster.ReloadComposeMetaValue()
	instSetAll := targetCluster.GetClusterValue().GetInstancesSetWhenSkipRouteFilter(true, true)
	if instSetAll.Count() > 0 {
		targetCluster.HasLimitedInstances = true
		return targetCluster, nil
	}
	defer targetCluster.PoolPut()
	return nil, errors.New("no available instances")
}

//不带金丝雀route
func (g *CanaryRouterFilter) noCanaryFilter(clusters model.ServiceClusters,
	withinCluster *model.Cluster) (*servicerouter.RouteResult, error) {
	availableCluster, err := g.noCanaryAvailableFilter(clusters, withinCluster)
	if err == nil && availableCluster != nil {
		result := servicerouter.PoolGetRouteResult(g.valueCtx)
		result.OutputCluster = availableCluster
		return result, nil
	}
	limitedCluster, err := g.noCanaryLimitedFilter(clusters, withinCluster)
	if err == nil && limitedCluster != nil {
		result := servicerouter.PoolGetRouteResult(g.valueCtx)
		result.OutputCluster = limitedCluster
		return result, nil
	}
	return nil, errors.New("no instances after canaryFilter")
}

//不带金丝雀过滤可用实例
func (g *CanaryRouterFilter) noCanaryAvailableFilter(clusters model.ServiceClusters,
	withinCluster *model.Cluster) (*model.Cluster, error) {
	notContainMetaKeyCluster := model.NewCluster(clusters, withinCluster)
	notContainMetaKeyCluster.AddMetadata(model.CanaryMetaKey, "")
	notContainMetaKeyCluster.ReloadComposeMetaValue()
	noTargetInstSet := notContainMetaKeyCluster.GetNotContainMetaKeyClusterValue().GetInstancesSet(false, true)
	if noTargetInstSet.Count() > 0 {
		return notContainMetaKeyCluster, nil
	}
	defer notContainMetaKeyCluster.PoolPut()

	// 优先返回带canary的可用实例
	containMetaKeyCluster := model.NewCluster(clusters, withinCluster)
	containMetaKeyCluster.AddMetadata(model.CanaryMetaKey, "")
	containMetaKeyCluster.ReloadComposeMetaValue()
	// 返回带canary的可用实例
	instSet := containMetaKeyCluster.GetContainMetaKeyClusterValue().GetInstancesSet(false, true)
	if instSet.Count() > 0 {
		return containMetaKeyCluster, nil
	}
	defer containMetaKeyCluster.PoolPut()
	return nil, errors.New("no available instances")
}

//不带金丝雀过滤limited实例
func (g *CanaryRouterFilter) noCanaryLimitedFilter(clusters model.ServiceClusters,
	withinCluster *model.Cluster) (*model.Cluster, error) {
	targetCluster := model.NewCluster(clusters, withinCluster)
	targetCluster.AddMetadata(model.CanaryMetaKey, "")
	targetCluster.ReloadComposeMetaValue()
	instSetAll := targetCluster.GetNotContainMetaKeyClusterValue().GetInstancesSetWhenSkipRouteFilter(true, true)
	if instSetAll.Count() > 0 {
		targetCluster.HasLimitedInstances = true
		return targetCluster, nil
	}
	defer targetCluster.PoolPut()
	return nil, errors.New("no available instances")
}

//是否需要启动规则路由
func (g *CanaryRouterFilter) Enable(routeInfo *servicerouter.RouteInfo) bool {
	return true
}

//init 注册插件
func init() {
	plugin.RegisterPlugin(&CanaryRouterFilter{})
}
