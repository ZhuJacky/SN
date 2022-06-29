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

package startup

import (
	"git.code.oa.com/polaris/polaris-go/pkg/config"
	"git.code.oa.com/polaris/polaris-go/pkg/flow/data"
	"git.code.oa.com/polaris/polaris-go/pkg/log"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	namingpb "git.code.oa.com/polaris/polaris-go/pkg/model/pb/v1"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/localregistry"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/serverconnector"
	"git.code.oa.com/polaris/polaris-go/pkg/version"
	"github.com/golang/protobuf/proto"
	"time"
)

//创建上报回调
func NewReportClientCallBack(
	cfg config.Configuration, supplier plugin.Supplier, globalCtx model.ValueContext) (*ReportClientCallBack, error) {
	var err error
	var callback = &ReportClientCallBack{}
	if callback.connector, err = data.GetServerConnector(cfg, supplier); nil != err {
		return nil, err
	}
	if callback.registry, err = data.GetRegistry(cfg, supplier); nil != err {
		return nil, err
	}
	callback.configuration = cfg
	callback.globalCtx = globalCtx
	callback.interval = cfg.GetGlobal().GetAPI().GetReportInterval()
	callback.loadLocalClientReportResult()
	return callback, nil
}

//上报客户端状态任务回调
type ReportClientCallBack struct {
	connector     serverconnector.ServerConnector
	registry      localregistry.InstancesRegistry
	configuration config.Configuration
	globalCtx     model.ValueContext
	interval      time.Duration
}

const (
	//地域信息持久化数据
	clientInfoPersistFile = "client_info.json"
)

//从本地缓存加载上报结果信息
func (r *ReportClientCallBack) loadLocalClientReportResult() {
	resp := &namingpb.Response{}
	cachedFile := clientInfoPersistFile
	err := r.registry.LoadPersistedMessage(cachedFile, resp)
	if nil != err {
		log.GetBaseLogger().Warnf("fail to load local region info from %s, err is %v", cachedFile, err)
		return
	}
	location := resp.GetClient().GetLocation()
	r.updateLocation(&model.Location{
		Region: location.GetRegion().GetValue(),
		Zone:   location.GetZone().GetValue(),
		Campus: location.GetCampus().GetValue(),
	}, nil)
}

//客户端上报的请求
func (r *ReportClientCallBack) reportClientRequest() *model.ReportClientRequest {
	apiConfig := r.configuration.GetGlobal().GetAPI()
	clientHost := apiConfig.GetBindIP()
	reportClientReq := &model.ReportClientRequest{
		Version: version.Version,
		Timeout: r.configuration.GetGlobal().GetAPI().GetTimeout(),
		PersistHandler: func(message proto.Message) error {
			return r.registry.PersistMessage(clientInfoPersistFile, message)
		},
	}
	if len(clientHost) > 0 {
		reportClientReq.Host = clientHost
	}
	return reportClientReq
}

//执行任务
func (r *ReportClientCallBack) Process(
	taskKey interface{}, taskValue interface{}, lastProcessTime time.Time) model.TaskResult {
	if !lastProcessTime.IsZero() && time.Since(lastProcessTime) < r.interval {
		return model.SKIP
	}
	svcKey := &model.ServiceKey{
		Service:   config.ServerDiscoverService,
		Namespace: config.ServerNamespace,
	}
	param := &model.ControlParam{}
	data.BuildControlParam(nil, r.configuration, param)
	reportClientReq := r.reportClientRequest()
	if err := reportClientReq.Validate(); nil != err {
		log.GetBaseLogger().Errorf("report client request fatal validate error:%v", err)
		return model.TERMINATE
	}
	resp, err := data.RetrySyncCall("clientReport",
		svcKey, reportClientReq, func(request interface{}) (interface{}, error) {
			return r.connector.ReportClient(request.(*model.ReportClientRequest))
		}, param)
	if nil != err {
		log.GetBaseLogger().Errorf("report client info:%+v, error:%v", reportClientReq, err)
		r.updateLocation(nil, err)
		//发生错误也要重试，直到获取到地域信息为止
		return model.CONTINUE
	}
	reportClientResp := resp.(*model.ReportClientResponse)
	r.updateLocation(&model.Location{
		Region: reportClientResp.Region,
		Zone:   reportClientResp.Zone,
		Campus: reportClientResp.Campus,
	}, nil)
	return model.CONTINUE
}

//更新区域属性
func (r *ReportClientCallBack) updateLocation(location *model.Location, lastErr model.SDKError) {
	if nil != location {
		//已获取到客户端的地域信息，更新到全局上下文
		log.GetBaseLogger().Infof("current client area info is {Region:%s, Zone:%s, Campus:%s}",
			location.Region, location.Zone, location.Campus)
	}
	if r.globalCtx.SetCurrentLocation(location, lastErr) {
		log.GetBaseLogger().Infof("client area info is ready")
	}
}
