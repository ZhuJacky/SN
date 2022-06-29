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

package serviceroute

import (
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/servicerouter"
	"sync"
	"sync/atomic"
)

//路由统计数据的key
type routeStatKey struct {
	plugId          int32
	retCode         model.ErrCode
}

//路由规则的key，标识一种路由规则
type ruleKey struct {
	plugId          int32
}

//存储服务路由数据的结构
type routeStatData struct {
	store *sync.Map
}

//添加一个服务路由的统计数据
func (r *routeStatData) putNewStat(gauge *servicerouter.RouteGauge) {
	key := routeStatKey{
		plugId:   gauge.PluginID,
		retCode:  gauge.RetCode,
	}
	dataInf, ok := r.store.Load(key)
	if !ok {
		var value uint32
		dataInf, _ = r.store.LoadOrStore(key, &value)
	}
	data := dataInf.(*uint32)
	atomic.AddUint32(data, 1)
}

//获取一个服务下面的所有路由规则的统计数据
func (r *routeStatData) getRouteRecord() map[ruleKey]map[model.ErrCode]uint32 {
	pluginRecordMap := make(map[ruleKey]map[model.ErrCode]uint32)
	r.store.Range(func(k, v interface{}) bool {
		value := v.(*uint32)
		num := atomic.LoadUint32(value)
		if num == 0 {
			return true
		}
		key := k.(routeStatKey)
		ruleKey := ruleKey{
			plugId:          key.plugId,
		}
		atomic.AddUint32(value, ^(num - 1))
		ruleMap, ok := pluginRecordMap[ruleKey]
		if !ok {
			ruleMap = make(map[model.ErrCode]uint32)
			pluginRecordMap[ruleKey] = ruleMap
		}
		ruleMap[key.retCode] = num
		return true
	})
	return pluginRecordMap
}
