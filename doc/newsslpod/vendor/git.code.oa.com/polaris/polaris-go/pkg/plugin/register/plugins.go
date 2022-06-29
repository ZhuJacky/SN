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

package register

import (
	//注册插件类型
	_ "git.code.oa.com/polaris/polaris-go/pkg/plugin/alarmreporter"
	_ "git.code.oa.com/polaris/polaris-go/pkg/plugin/circuitbreaker"
	_ "git.code.oa.com/polaris/polaris-go/pkg/plugin/loadbalancer"
	_ "git.code.oa.com/polaris/polaris-go/pkg/plugin/localregistry"
	_ "git.code.oa.com/polaris/polaris-go/pkg/plugin/outlierdetection"
	_ "git.code.oa.com/polaris/polaris-go/pkg/plugin/ratelimiter"
	_ "git.code.oa.com/polaris/polaris-go/pkg/plugin/serverconnector"
	_ "git.code.oa.com/polaris/polaris-go/pkg/plugin/servicerouter"
	_ "git.code.oa.com/polaris/polaris-go/pkg/plugin/statreporter"
	_ "git.code.oa.com/polaris/polaris-go/pkg/plugin/subscribe"
	_ "git.code.oa.com/polaris/polaris-go/pkg/plugin/weightadjuster"

	//注册具体插件实例
	_ "git.code.oa.com/polaris/polaris-go/plugin/alarmreporter/file"
	_ "git.code.oa.com/polaris/polaris-go/plugin/circuitbreaker/errorcount"
	_ "git.code.oa.com/polaris/polaris-go/plugin/circuitbreaker/errorrate"
	_ "git.code.oa.com/polaris/polaris-go/plugin/loadbalancer/hash"
	_ "git.code.oa.com/polaris/polaris-go/plugin/loadbalancer/maglev"
	_ "git.code.oa.com/polaris/polaris-go/plugin/loadbalancer/ringhash"
	_ "git.code.oa.com/polaris/polaris-go/plugin/loadbalancer/weightedrandom"
	_ "git.code.oa.com/polaris/polaris-go/plugin/localregistry/inmemory"
	_ "git.code.oa.com/polaris/polaris-go/plugin/logger/zaplog"
	_ "git.code.oa.com/polaris/polaris-go/plugin/outlierdetection/http"
	_ "git.code.oa.com/polaris/polaris-go/plugin/outlierdetection/tcp"
	_ "git.code.oa.com/polaris/polaris-go/plugin/outlierdetection/udp"
	_ "git.code.oa.com/polaris/polaris-go/plugin/ratelimiter/reject"
	_ "git.code.oa.com/polaris/polaris-go/plugin/ratelimiter/unirate"
	_ "git.code.oa.com/polaris/polaris-go/plugin/serverconnector/grpc"

	_ "git.code.oa.com/polaris/polaris-go/plugin/servicerouter/canary"
	_ "git.code.oa.com/polaris/polaris-go/plugin/servicerouter/dstmeta"
	_ "git.code.oa.com/polaris/polaris-go/plugin/servicerouter/filteronly"
	_ "git.code.oa.com/polaris/polaris-go/plugin/servicerouter/nearbybase"
	_ "git.code.oa.com/polaris/polaris-go/plugin/servicerouter/rulebase"
	_ "git.code.oa.com/polaris/polaris-go/plugin/servicerouter/setdivision"

	_ "git.code.oa.com/polaris/polaris-go/plugin/statreporter/monitor"
	_ "git.code.oa.com/polaris/polaris-go/plugin/statreporter/ratelimit"
	_ "git.code.oa.com/polaris/polaris-go/plugin/statreporter/serviceinfo"
	_ "git.code.oa.com/polaris/polaris-go/plugin/statreporter/serviceroute"
	_ "git.code.oa.com/polaris/polaris-go/plugin/subscribe/localchannel"
	_ "git.code.oa.com/polaris/polaris-go/plugin/weightadjuster/ratedelay"
)
