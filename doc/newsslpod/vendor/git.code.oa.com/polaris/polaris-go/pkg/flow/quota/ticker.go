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
	"git.code.oa.com/polaris/polaris-go/pkg/plugin"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/localregistry"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/serverconnector"
	"sync/atomic"
	"time"
)

//远程配额查询任务
type RemoteQuotaCallBack struct {
	registry             localregistry.InstancesRegistry
	rLimitConnector      serverconnector.RateLimitConnector
	asyncRLimitConnector serverconnector.AsyncRateLimitConnector
	engine               model.Engine
}

//创建查询任务
func NewRemoteQuotaCallback(cfg config.Configuration, supplier plugin.Supplier,
	engine model.Engine) (*RemoteQuotaCallBack, error) {
	registry, err := data.GetRegistry(cfg, supplier)
	if nil != err {
		return nil, err
	}
	connector, err := data.GetServerConnector(cfg, supplier)
	if nil != err {
		return nil, err
	}
	return &RemoteQuotaCallBack{
		registry:             registry,
		rLimitConnector:      connector.GetRateLimitConnector(),
		asyncRLimitConnector: connector.GetAsyncRateLimitConnector(),
		engine:               engine}, nil
}

//处理远程配额查询任务
func (r *RemoteQuotaCallBack) Process(
	taskKey interface{}, taskValue interface{}, lastProcessTime time.Time) model.TaskResult {
	rateLimitWindow := taskValue.(*RateLimitWindow)
	now := clock.GetClock().Now()
	if !lastProcessTime.IsZero() && now.Sub(lastProcessTime) < rateLimitWindow.syncParam.reportInterval {
		if !atomic.CompareAndSwapInt32(&rateLimitWindow.syncParam.syncFlag, SyncAtOnce, SyncInterval) {
			//判断是否实时上报
			return model.SKIP
		}
	}
	lastVisitTime := rateLimitWindow.GetLastQuotaAccessTime()
	lastStatus := rateLimitWindow.GetStatus()
	if lastStatus == Expired || lastStatus == Deleted {
		r.asyncRLimitConnector.ClearExpireWindow(rateLimitWindow.quotaKey)
		deleteWindow(rateLimitWindow.WindowSet, rateLimitWindow.WindowSetKey)
		return model.TERMINATE
	}
	if now.Sub(lastVisitTime) >= rateLimitWindow.expireDuration {
		log.GetBaseLogger().Infof(
			"quota(key=%s, service=%s) has been expired, expire time is %v",
			rateLimitWindow.quotaKey, rateLimitWindow.SvcKey, now)
		rateLimitWindow.SetStatus(Expired)
		return model.CONTINUE
	}
	//状态机
	switch rateLimitWindow.GetStatus() {
	case Created:
		break
	case Initializing:
		break
	case RemoteInitFail:
		go rateLimitWindow.doRemoteInitialize(r.rLimitConnector)
	case RemoteAcquireFail:
		go rateLimitWindow.doRemoteInitialize(r.rLimitConnector)
	default:
		rateLimitWindow.doAsyncRemoteAcquire(r.asyncRLimitConnector)
	}
	return model.CONTINUE
}
