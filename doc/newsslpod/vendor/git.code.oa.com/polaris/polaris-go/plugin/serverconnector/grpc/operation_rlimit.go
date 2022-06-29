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

package grpc

import (
	"bytes"
	"fmt"
	"git.code.oa.com/polaris/polaris-go/pkg/clock"
	"git.code.oa.com/polaris/polaris-go/pkg/config"
	"git.code.oa.com/polaris/polaris-go/pkg/log"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	rlimitpb "git.code.oa.com/polaris/polaris-go/pkg/model/pb/metric"
	"git.code.oa.com/polaris/polaris-go/pkg/network"
	connector "git.code.oa.com/polaris/polaris-go/plugin/serverconnector/common"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"sync/atomic"
	"time"
)

var (
	MetricMsgId int64 = 0
)

//限流Server连接器
type rateLimitConnector struct {
	connManager network.ConnectionManager
}

func GetNextMsgId() int64 {
	msgId := atomic.AddInt64(&MetricMsgId, 1)
	return msgId
}

//解析限流请求
func parseRateLimitRequest(request proto.Message) (*rlimitpb.RateLimitRequest, []byte) {
	rlReq := request.(*rlimitpb.RateLimitRequest)
	hashKey := make([]byte, 0, len(rlReq.GetKey().GetValue()))
	buf := bytes.NewBuffer(hashKey)
	buf.WriteString(rlReq.GetKey().GetValue())
	return rlReq, buf.Bytes()
}

func formatMetricHashKey(metricKey *rlimitpb.MetricKey) []byte {
	if metricKey == nil {
		return nil
	}
	hashStr := metricKey.GetNamespace() + config.DefaultNamesSeparator + metricKey.GetService() +
		config.DefaultNamesSeparator + metricKey.Labels
	return []byte(hashStr)
}

//初始化限流控制信息
func (r *rateLimitConnector) Initialize(request proto.Message, timeout time.Duration) (proto.Message, error) {
	rlReq, hashKey := parseRateLimitRequest(request)
	opKey := connector.OpKeyRateLimitInit
	startTime := clock.GetClock().Now()
	conn, err := r.connManager.GetConnectionByHashKey(opKey, config.MetricCluster, hashKey)
	if nil != err {
		return nil, connector.NetworkError(r.connManager, conn, int32(model.ErrCodeConnectError), err, startTime,
			fmt.Sprintf("fail to get connection, opKey %s", opKey))
	}
	defer conn.Release(opKey)
	rlimitClient := rlimitpb.NewRateLimitGRPCClient(network.ToGRPCConn(conn.Conn))
	reqID := connector.NextRateLimitInitReqID()
	ctx, cancel := connector.CreateHeaderContext(timeout, reqID)
	if cancel != nil {
		defer cancel()
	}
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		reqJson, _ := (&jsonpb.Marshaler{}).MarshalToString(rlReq)
		log.GetBaseLogger().Debugf("[RateLimit]----------Initialize:  %s", conn.Address)
		log.GetBaseLogger().Debugf("request to send is %s, opKey %s, connID %s", reqJson, opKey, conn.ConnID)
	}
	resp, err := rlimitClient.InitializeQuota(ctx, rlReq)
	if nil != err {
		return nil, connector.NetworkError(r.connManager, conn, int32(model.ErrorCodeRpcError), err, startTime,
			fmt.Sprintf("fail to send request, opKey %s, reqID %s, connID %s", opKey, reqID, conn.ConnID))
	}
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		respJson, _ := (&jsonpb.Marshaler{}).MarshalToString(resp)
		log.GetBaseLogger().Debugf("response recv is %s, opKey %s, connID %s", respJson, opKey, conn.ConnID)
	}
	endTime := clock.GetClock().Now()
	r.connManager.ReportSuccess(conn.ConnID, connector.GetConnErrorCode(nil), endTime.Sub(startTime))
	return r.checkRespError(resp, opKey, reqID, conn)
}

func (r *rateLimitConnector) checkRespError(
	resp *rlimitpb.RateLimitResponse, opKey string, reqID string, conn *network.Connection) (proto.Message, error) {
	if !model.IsSuccessResultCode(resp.GetCode().GetValue()) {
		if model.IsServerException(resp.GetCode().GetValue()) {
			return nil, model.NewSDKError(model.ErrCodeServerException, nil,
				"server exception from (opKey %s, reqID %s, connID %s), err is (code %d, info %s)",
				opKey, reqID, conn.ConnID, resp.GetCode().GetValue(), resp.GetInfo().GetValue())
		}
		return nil, model.NewSDKError(model.ErrCodeServerUserError, nil,
			"client exception from (opKey %s, reqID %s, connID %s), err code %d, info %s",
			opKey, reqID, conn.ConnID, resp.GetCode().GetValue(), resp.GetInfo().GetValue())

	}
	return resp, nil
}

//上报并获取分布式限流配额
func (r *rateLimitConnector) Acquire(request proto.Message, timeout time.Duration) (proto.Message, error) {
	rlReq, hashKey := parseRateLimitRequest(request)
	opKey := connector.OpKeyRateLimitAcquire
	startTime := clock.GetClock().Now()
	conn, err := r.connManager.GetConnectionByHashKey(opKey, config.MetricCluster, hashKey)

	if nil != err {
		return nil, connector.NetworkError(r.connManager, conn, int32(model.ErrCodeConnectError), err, startTime,
			fmt.Sprintf("fail to get connection, opKey %s", opKey))
	}
	defer conn.Release(opKey)
	rlimitClient := rlimitpb.NewRateLimitGRPCClient(network.ToGRPCConn(conn.Conn))
	reqID := connector.NextRateLimitAcquireReqID()
	ctx, cancel := connector.CreateHeaderContext(timeout, reqID)
	if cancel != nil {
		defer cancel()
	}
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		reqJson, _ := (&jsonpb.Marshaler{}).MarshalToString(rlReq)
		log.GetBaseLogger().Debugf("request to send is %s, opKey %s, connID %s", reqJson, opKey, conn.ConnID)
	}
	streamClient, err := rlimitClient.AcquireQuota(ctx)
	if nil != err {
		return nil, connector.NetworkError(r.connManager, conn, int32(model.ErrCodeConnectError), err, startTime,
			fmt.Sprintf("fail to create stream, opKey %s, connID %s", opKey, conn.ConnID))
	}
	defer streamClient.CloseSend()
	err = streamClient.Send(rlReq)
	if nil != err {
		return nil, connector.NetworkError(r.connManager, conn, int32(model.ErrCodeConnectError), err, startTime,
			fmt.Sprintf("fail to send request, opKey %s, reqID %s, connID %s", opKey, reqID, conn.ConnID))
	}
	resp, err := streamClient.Recv()
	if nil != err {
		return nil, connector.NetworkError(r.connManager, conn, int32(model.ErrorCodeRpcError), err, startTime,
			fmt.Sprintf("fail to recv response, opKey %s, reqID %s, connID %s", opKey, reqID, conn.ConnID))
	}
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		respJson, _ := (&jsonpb.Marshaler{}).MarshalToString(resp)
		log.GetBaseLogger().Debugf("response recv is %s, opKey %s, connID %s", respJson, opKey, conn.ConnID)
	}
	endTime := clock.GetClock().Now()
	r.connManager.ReportSuccess(conn.ConnID, connector.GetConnErrorCode(nil), endTime.Sub(startTime))
	return r.checkRespError(resp, opKey, reqID, conn)
}

func checkRspErrBySpecificInfo(code uint32, info string, opKey string, msgId int64, connId string) error {
	if !model.IsSuccessResultCode(code) {
		if model.IsServerException(code) {
			return model.NewSDKError(model.ErrCodeServerException, nil,
				"server exception from (opKey %s, msgID %d, connID %s), err is (code %d, info %s)",
				opKey, msgId, connId, code, info)
		}
		return model.NewSDKError(model.ErrCodeServerUserError, nil,
			"client exception from (opKey %s, msgID %d, connID %s), err code %d, info %s",
			opKey, msgId, connId, code, info)
	}
	return nil
}

func (r *rateLimitConnector) Init(request proto.Message, timeout time.Duration) (proto.Message, error) {
	metricReq := request.(*rlimitpb.MetricInitRequest)
	hashKey := formatMetricHashKey(metricReq.GetKey())
	if hashKey == nil {
		return nil, model.NewSDKError(model.ErrCodeInternalError, nil, "")
	}
	metricReq.MsgId = &wrappers.Int64Value{Value: GetNextMsgId()}
	opKey := connector.OpKeyRateLimitMetricInit
	startTime := clock.GetClock().Now()
	conn, err := r.connManager.GetConnectionByHashKey(opKey, config.MetricCluster, hashKey)
	if nil != err {
		return nil, connector.NetworkError(r.connManager, conn, int32(model.ErrCodeConnectError), err, startTime,
			fmt.Sprintf("fail to get connection, opKey %s", opKey))
	}
	defer conn.Release(opKey)
	metricClient := rlimitpb.NewMetricGRPCClient(network.ToGRPCConn(conn.Conn))
	reqID := connector.NextRateLimitAcquireReqID()
	ctx, cancel := connector.CreateHeaderContext(timeout, reqID)
	if cancel != nil {
		defer cancel()
	}
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		reqJson, _ := (&jsonpb.Marshaler{}).MarshalToString(metricReq)
		log.GetBaseLogger().Debugf("request to send is %s, opKey %s, connID %s", reqJson, opKey, conn.ConnID)
	}
	resp, err := metricClient.Init(ctx, metricReq)
	if nil != err {
		return nil, connector.NetworkError(r.connManager, conn, int32(model.ErrorCodeRpcError), err, startTime,
			fmt.Sprintf("fail to send request, opKey %s, reqID %s, connID %s", opKey, reqID, conn.ConnID))
	}
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		respJson, _ := (&jsonpb.Marshaler{}).MarshalToString(resp)
		log.GetBaseLogger().Debugf("response recv is %s, opKey %s, connID %s", respJson, opKey, conn.ConnID)
	}
	endTime := clock.GetClock().Now()
	_ = endTime
	err = checkRspErrBySpecificInfo(resp.GetCode().Value, resp.GetInfo().Value, opKey, resp.GetMsgId().Value,
		conn.ConnID.String())
	return resp, err
}

//上报到metric-server, 用于监控
func (r *rateLimitConnector) Report(request proto.Message, timeout time.Duration) (proto.Message, error) {
	metricReq := request.(*rlimitpb.MetricRequest)
	hashKey := formatMetricHashKey(metricReq.GetKey())
	if hashKey == nil {
		return nil, model.NewSDKError(model.ErrCodeInternalError, nil, "")
	}
	metricReq.MsgId = &wrappers.Int64Value{Value: GetNextMsgId()}
	opKey := connector.OpKeyRateLimitMetricReport
	startTime := clock.GetClock().Now()
	conn, err := r.connManager.GetConnectionByHashKey(opKey, config.MetricCluster, hashKey)
	if nil != err {
		return nil, connector.NetworkError(r.connManager, conn, int32(model.ErrCodeConnectError), err, startTime,
			fmt.Sprintf("fail to get connection, opKey %s", opKey))
	}
	defer conn.Release(opKey)
	metricClient := rlimitpb.NewMetricGRPCClient(network.ToGRPCConn(conn.Conn))
	reqID := connector.NextRateLimitAcquireReqID()
	ctx, cancel := connector.CreateHeaderContext(timeout, reqID)
	if cancel != nil {
		defer cancel()
	}
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		reqJson, _ := (&jsonpb.Marshaler{}).MarshalToString(metricReq)
		log.GetBaseLogger().Debugf("request to send is %s, opKey %s, connID %s", reqJson, opKey, conn.ConnID)
	}
	streamClient, err := metricClient.Report(ctx)
	if nil != err {
		return nil, connector.NetworkError(r.connManager, conn, int32(model.ErrCodeConnectError), err, startTime,
			fmt.Sprintf("fail to send request, opKey %s, reqID %s, connID %s", opKey, reqID, conn.ConnID))
	}
	defer streamClient.CloseSend()
	err = streamClient.Send(metricReq)
	if nil != err {
		return nil, connector.NetworkError(r.connManager, conn, int32(model.ErrorCodeRpcError), err, startTime,
			fmt.Sprintf("fail to send request, opKey %s, reqID %s, connID %s", opKey, reqID, conn.ConnID))
	}
	resp, err := streamClient.Recv()
	if nil != err {
		return nil, connector.NetworkError(r.connManager, conn, int32(model.ErrorCodeRpcError), err, startTime,
			fmt.Sprintf("fail to recv response, opKey %s, reqID %s, connID %s", opKey, reqID, conn.ConnID))
	}
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		respJson, _ := (&jsonpb.Marshaler{}).MarshalToString(resp)
		log.GetBaseLogger().Debugf("response recv is %s, opKey %s, connID %s", respJson, opKey, conn.ConnID)
	}
	endTime := clock.GetClock().Now()
	_ = endTime
	err = checkRspErrBySpecificInfo(resp.GetCode().Value, resp.GetInfo().Value, opKey, resp.GetMsgId().Value,
		conn.ConnID.String())
	return resp, err
}
