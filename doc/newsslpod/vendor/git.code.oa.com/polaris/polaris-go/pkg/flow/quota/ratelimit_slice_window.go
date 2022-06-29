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
	"git.code.oa.com/polaris/polaris-go/pkg/metric"
	"sync/atomic"
	"time"
)

type RateLimitSliceWindow struct {
	*metric.SliceWindow

	window *RateLimitWindow
	//用于上报计数，周期清0
	limited uint32
	//拥有上报pass的计算，周期清0
	passed uint32
}

func NewRateLimitSliceWindow(typ string, bucketCount int, bucketInterval time.Duration, metricSize int,
	curTime int64, rlWin *RateLimitWindow) *RateLimitSliceWindow {
	window := &RateLimitSliceWindow{}
	window.SliceWindow = metric.NewSliceWindow(typ, bucketCount, bucketInterval, metricSize, curTime)
	window.limited = 0
	window.passed = 0
	window.window = rlWin
	return window
}

func (rs *RateLimitSliceWindow) AddAcquireInfo(curTime int64, passed bool) {
	startTime := rs.CalcStartTime(curTime)
	if atomic.LoadInt64(&rs.PeriodStartTime) == startTime {
		if passed {
			atomic.AddUint32(&rs.passed, 1)
		} else {
			atomic.AddUint32(&rs.limited, 1)
		}
		return
	} else {
		rs.Lock.Lock()
		defer rs.Lock.Unlock()
		nowPeriodStartTime := atomic.LoadInt64(&rs.PeriodStartTime)
		if nowPeriodStartTime == startTime {
			if passed {
				atomic.AddUint32(&rs.passed, 1)
			} else {
				atomic.AddUint32(&rs.limited, 1)
			}
			return
		} else {
			atomic.StoreInt64(&rs.PeriodStartTime, startTime)
			if passed {
				atomic.SwapUint32(&rs.passed, 1)
			} else {
				atomic.SwapUint32(&rs.limited, 1)
			}
		}
	}
}

func (rs *RateLimitSliceWindow) GetLimited(curTime int64) uint32 {
	startTime := rs.CalcStartTime(curTime)
	var retNum uint32 = 0
	if atomic.LoadInt64(&rs.PeriodStartTime) == startTime {
		retNum = atomic.SwapUint32(&rs.limited, 0)
		return retNum
	} else {
		return 0
	}
}

func (rs *RateLimitSliceWindow) GetPassed(curTime int64) uint32 {
	startTime := rs.CalcStartTime(curTime)
	var retNum uint32 = 0
	if atomic.LoadInt64(&rs.PeriodStartTime) == startTime {
		retNum = atomic.SwapUint32(&rs.passed, 0)
		return retNum
	} else {
		return 0
	}
}
