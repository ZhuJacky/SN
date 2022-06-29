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
	"git.code.oa.com/polaris/polaris-go/pkg/log"
	"git.code.oa.com/polaris/polaris-go/pkg/metric"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	"git.code.oa.com/polaris/polaris-go/pkg/model/pb"
	namingpb "git.code.oa.com/polaris/polaris-go/pkg/model/pb/v1"
	"math"
	"sort"
	"sync/atomic"
	"time"
)

const (
	//滑窗类型名
	quotaUsageMetricType = "quotaUsage"
	//滑窗的桶数量
	metricBucketCount = 1
)

const (
	//滑窗统计值标签
	dimensionQuotaCount = iota
	//最大标签范围
	dimensionMax
)

// 创建QPS远程限流窗口
func NewRemoteAwareQpsBucket(window *RateLimitWindow, rule *namingpb.Rule, instCount int32) *RemoteAwareQpsBucket {
	raqb := &RemoteAwareQpsBucket{
		window:    window,
		instCount: instCount,
		mode:      rule.GetType(),
	}
	if raqb.instCount == 0 {
		raqb.instCount = 1
	}
	raqb.reportAmountPresent = float64(100-rule.GetReport().GetAmountPercent().GetValue()) / float64(100)
	raqb.tokenBuckets = initTokenBuckets(raqb.mode, rule.Amounts, instCount, window)
	raqb.tokenBucketMap = make(map[int64]*TokenBucket, len(raqb.tokenBuckets))

	var minDuration int64
	for _, tokenBucket := range raqb.tokenBuckets {
		validDuration := tokenBucket.validDuration
		if minDuration == 0 || minDuration > validDuration {
			minDuration = validDuration
		}
		raqb.tokenBucketMap[validDuration] = tokenBucket
	}
	//远程请求最大超时时间为最小周期+1s
	raqb.maxRemoteWait = minDuration + 1000
	raqb.lastRemoteUpdateTime = 0
	return raqb
}

//远程下发的配额值
type quotaUsageInfo struct {
	//等待上报的
	pendingReportQuotas map[int64]*metric.SliceWindow
	//当前正在使用的配额
	currentQuotaUsed int64
}

// 远程配额分配的算法桶
type RemoteAwareQpsBucket struct {
	//所属的限流窗口
	window *RateLimitWindow
	//最小周期，上报应答超过1个最小周期没有返回，则认为超时
	maxRemoteWait int64
	//实时上百百分比
	reportAmountPresent float64
	// 令牌桶
	tokenBuckets TokenBuckets
	// 令牌桶map，用于索引
	tokenBucketMap map[int64]*TokenBucket
	// 配额使用情况(真实使用情况)
	//quotaUsage quotaUsageInfo
	// 实例数量
	instCount int32
	//限流模式
	mode namingpb.Rule_Type
	// 上一次远程更新时间
	lastRemoteUpdateTime int64
}

// 执行配额分配操作
func (r *RemoteAwareQpsBucket) Allocate() *model.QuotaResponse {
	if len(r.tokenBuckets) == 0 {
		return &model.QuotaResponse{
			Code: model.QuotaResultOk,
			Info: "rule has no amount config",
		}
	}
	var stopIndex = -1
	now := clock.GetClock().Now()
	lastRemoteUpdateTime := atomic.LoadInt64(&r.lastRemoteUpdateTime)
	passedTimeAfterLastUpdate := now.UnixNano()/1e6 - lastRemoteUpdateTime
	remoteQuotaExpired := passedTimeAfterLastUpdate > r.maxRemoteWait
	//slowAcquire := atomic.LoadInt32(&r.window.continuousSlowRemoteAcquireCount) > 3
	usedRemoteQuota := r.mode == namingpb.Rule_GLOBAL && lastRemoteUpdateTime > 0 && !remoteQuotaExpired &&
		r.window.configMode == model.ConfigQuotaGlobalMode
	//是否需要立刻发起上报，只有配额扣除成功才执行立刻上报
	var reportAtOnce bool
	serverNow := now.UnixNano()/int64(time.Millisecond) + r.window.GetRemoteDiff()
	for i, tokenBucket := range r.tokenBuckets {
		//先增加配额到本地滑窗
		pass, report := r.tokenTryAllocate(tokenBucket, serverNow, usedRemoteQuota)
		if !pass {
			stopIndex = i
			break
		}
		if report {
			reportAtOnce = report
		}
	}
	if stopIndex < 0 {
		//分配成功
		if reportAtOnce {
			//发起实时上报
			atomic.CompareAndSwapInt32(&r.window.syncParam.syncFlag, SyncInterval, SyncAtOnce)
		}
		if r.window.configMode == model.ConfigQuotaGlobalMode {
			for _, token := range r.tokenBuckets {
				token.sliceWindow.AddAcquireInfo(serverNow, true)
			}
		}
		r.monitorReport(QuotaGranted, stopIndex)
		return &model.QuotaResponse{
			Code: model.QuotaResultOk,
		}
	}
	if r.window.configMode == model.ConfigQuotaGlobalMode {
		r.tokenBuckets[stopIndex].sliceWindow.AddAcquireInfo(serverNow, false)
	}
	//log.GetBaseLogger().Infof("limit")
	//分配失败，归还预扣除的配额
	for i := 0; i < stopIndex+1; i++ {
		bucket := r.tokenBuckets[i]
		//判断是否需要从远程配额中摘取
		if usedRemoteQuota {
			rQuotaInfo := bucket.getRemoteQuota()
			atomic.AddInt64(&rQuotaInfo.remoteTokenLeft, 1)
			rQuotaInfo.releaseWriter()
		}
		//归还配额给本地滑窗
		_ = bucket.sliceWindow.AddGaugeByValueByMillTime(-1, serverNow)
	}
	//上报
	r.monitorReport(QuotaLimited, stopIndex)
	return &model.QuotaResponse{
		Code: model.QuotaResultLimited,
	}
}

func (r *RemoteAwareQpsBucket) tokenTryAllocate(tokenBucket *TokenBucket, now int64,
	usedRemoteQuota bool) (bool, bool) {
	nextQps := tokenBucket.sliceWindow.AddGaugeByValueByMillTime(1, now)
	if usedRemoteQuota {
		rQuotaInfo := tokenBucket.getRemoteQuota()
		if nextQps == 1 {
			tokenTotal := atomic.LoadInt64(&tokenBucket.tokenLocal)
			rQuotaInfo.flushNewPeriodQuota(now, tokenBucket.validDuration, tokenTotal, r.reportAmountPresent)
		}
		restQuota := atomic.AddInt64(&rQuotaInfo.remoteTokenLeft, -1)
		rQuotaInfo.releaseWriter()
		if restQuota < 0 {
			// 远程配额检查不通过
			return false, false
		}
		if rQuotaInfo.reportThreshold > 0 && restQuota == rQuotaInfo.reportThreshold {
			log.GetBaseLogger().Debugf(
				"schedule report at once by quotaUsage %d, duration %v", restQuota, tokenBucket.validDuration)
			return true, true
		}
	} else {
		tokenTotal := atomic.LoadInt64(&tokenBucket.tokenLocal)
		if tokenTotal-nextQps < 0 {
			return false, false
		}
	}
	return true, false
}

func (r *RemoteAwareQpsBucket) monitorReport(limitType RateLimitType, stopIndex int) {
	gauge := &RateLimitGauge{
		EmptyInstanceGauge: model.EmptyInstanceGauge{},
		Window:             r.window,
		Type:               limitType,
		AmountIndex:        stopIndex,
	}
	r.window.engine.SyncReportStat(model.RateLimitStat, gauge)
}

// 执行配额回收操作
func (r *RemoteAwareQpsBucket) Release() {
	// 对于QPS限流，无需进行释放
}

// 设置通过限流服务端获取的远程QPS
func (r *RemoteAwareQpsBucket) SetRemoteQuota(remoteQuotas *RemoteQuotaResult) {
	instCount := atomic.LoadInt32(&r.instCount)
	currentUsage := remoteQuotas.CurrentUsage
	currentTime := clock.GetClock().Now()
	for _, limiter := range remoteQuotas.RemoteQuotas {
		goDuration, _ := pb.ConvertDuration(limiter.GetDuration())
		goDurationMilli := model.ToMilliSeconds(goDuration)
		tokenBucket, ok := r.tokenBucketMap[goDurationMilli]
		if !ok {
			continue
		}
		if log.GetBaseLogger().IsLevelEnabled(log.TraceLog) {
			log.GetBaseLogger().Tracef(
				"SetRemoteQuota %s %s", r.window.SvcKey.String(), limiter.String())
		}

		tokenLeftTotal := tokenBucket.tokenTotal - int64(limiter.GetAmount().GetValue())
		var tokenLeftInst float64
		if tokenLeftTotal > 0 {
			tokenLeftInst = math.Ceil(float64(tokenLeftTotal) / float64(instCount))
		}
		var quotaUsed int64
		if nil != currentUsage && len(currentUsage.QuotaUsed) > 0 {
			quotaUsed = int64(currentUsage.QuotaUsed[goDurationMilli])
		}

		rQuota := &remoteQuotaInfo{
			updating:             1,
			updateTime:           currentTime,
			remoteTokenTotal:     int64(tokenLeftInst),
			reportThreshold:      int64(r.reportAmountPresent * float64(tokenLeftInst)),
			lastUpdateServerTime: remoteQuotas.ServerTime,
		}
		lastQuota := tokenBucket.remoteQuota.Load().(*remoteQuotaInfo)
		tokenBucket.remoteQuota.Store(rQuota)
		lastQuota.waitUpdateFinish()
		if model.ToMilliSeconds(currentTime.Sub(lastQuota.updateTime)) >= goDurationMilli {
			//初始化或者配额已经过程，则丢弃老的配额，使用总配额作为可用配额
			rQuota.remoteTokenLeft = rQuota.remoteTokenTotal
		} else {
			//计算出在上报过程中用了多少配额
			quotaWhenAcquire := quotaUsed
			log.GetBaseLogger().Tracef(
				"SetRemoteQuota quotaWhenAcquire %d", quotaWhenAcquire)
			if quotaWhenAcquire > 0 {
				rQuota.remoteTokenLeft = rQuota.remoteTokenTotal - quotaWhenAcquire
			} else {
				rQuota.remoteTokenLeft = rQuota.remoteTokenTotal
			}
		}
		if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
			log.GetBaseLogger().Debugf("[RateLimitAcquire] %d SetRemoteQuota %s goDurationMilli:%d "+
				"remoteTokenTotal:%d remoteTokenLeft:%d reportThreshold:%d serverTime:%d",
				time.Now().UnixNano()/int64(time.Millisecond), r.window.SvcKey.String(),
				goDurationMilli, rQuota.remoteTokenTotal, rQuota.remoteTokenLeft, rQuota.reportThreshold,
				remoteQuotas.ServerTime)
		}
		//放开控制，允许流量分配
		atomic.StoreInt32(&rQuota.updating, 0)
	}
	now := clock.GetClock().Now()
	atomic.StoreInt64(&r.lastRemoteUpdateTime, now.UnixNano()/1e6)
}

// 获取已使用的配额
func (r *RemoteAwareQpsBucket) GetQuotaUsedForAcquire(nowTime int64) *UsageInfo {
	//上报时候进行清零
	result := &UsageInfo{
		QuotaUsed: make(map[int64]uint32, len(r.tokenBuckets)),
		CurTime:   nowTime,
		Limited:   make(map[int64]uint32, len(r.tokenBuckets)),
	}
	for key, tokenBucket := range r.tokenBucketMap {
		usedNum := tokenBucket.sliceWindow.GetPassed(nowTime)
		limitedNum := tokenBucket.sliceWindow.GetLimited(nowTime)
		result.QuotaUsed[key] = usedNum
		result.Limited[key] = limitedNum
	}
	return result
}

func (r *RemoteAwareQpsBucket) GetQuotaUsed(nowTime int64) *UsageInfo {
	result := &UsageInfo{
		QuotaUsed: make(map[int64]uint32, len(r.tokenBuckets)),
		CurTime:   nowTime,
	}
	for key, tokenBucket := range r.tokenBucketMap {
		usedNum := tokenBucket.sliceWindow.GetPassed(nowTime)
		result.QuotaUsed[key] = uint32(usedNum)
	}
	return result
}

//更新服务实例个数
func (r *RemoteAwareQpsBucket) UpdateInstanceCount(count int32) {
	if count == 0 {
		//防止除零异常
		count = 1
	}
	atomic.StoreInt32(&r.instCount, count)
	for _, bucket := range r.tokenBuckets {
		tokenPerInst := getTokenPerInst(r.mode, bucket.tokenTotal, count)
		atomic.StoreInt64(&bucket.tokenLocal, tokenPerInst)
	}
}

//初始化sliceWindow周期开始时间（和server开始窗口对齐）
func (r *RemoteAwareQpsBucket) InitPeriodStart(now int64) {
	for _, t := range r.tokenBuckets {
		t.sliceWindow.SetPeriodStart(now)
	}
}

//GetMaxRemoteWait 获取最小周期
func (r *RemoteAwareQpsBucket) GetMaxRemoteWait() int64 {
	return r.maxRemoteWait
}

// 远程配额信息
type remoteQuotaInfo struct {
	//配额下发时间
	updateTime time.Time
	//用了多少个以后开始上报
	reportThreshold int64
	// 远程剩余配额，用于划扣
	remoteTokenLeft int64
	// 远程下发的配额数，常量，不用于划扣
	remoteTokenTotal int64
	//依赖计数
	writerCount int32
	//是否正在更新
	updating int32
	//上次更新,server的时间, 单位ms
	lastUpdateServerTime int64
}

// 占据
func (r *remoteQuotaInfo) occupyWriter() {
	atomic.AddInt32(&r.writerCount, 1)
}

// 释放
func (r *remoteQuotaInfo) releaseWriter() {
	atomic.AddInt32(&r.writerCount, -1)
}

// CAS等待依赖计数器清零
func (r *remoteQuotaInfo) waitUpdateFinish() {
	for {
		if atomic.LoadInt32(&r.writerCount) <= 0 {
			return
		}
	}
}

//刷新配额
func (r *remoteQuotaInfo) flushNewPeriodQuota(curServerTime int64, durationMill int64, quotaLocal int64,
	reportAmountPresent float64) bool {
	startTime := curServerTime - curServerTime%durationMill
	lastUpdateSvrTime := atomic.LoadInt64(&r.lastUpdateServerTime)
	loadStart := lastUpdateSvrTime - lastUpdateSvrTime%durationMill
	if startTime == loadStart {
		log.GetBaseLogger().Debugf("remoteQuotaInfo no need to flushNewPeriodQuota %d %d", startTime, loadStart)
		return false
	} else {
		log.GetBaseLogger().Debugf("remoteQuotaInfo flushNewPeriodQuota %d %d", startTime, loadStart)
		atomic.StoreInt64(&r.remoteTokenLeft, quotaLocal)
		atomic.StoreInt64(&r.reportThreshold, int64(reportAmountPresent*float64(quotaLocal)))
		return true
	}
}

// 令牌桶
type TokenBucket struct {
	// 限流区间
	validDuration int64
	// 每周期分配的配额总量
	tokenTotal int64
	// 统计滑窗
	sliceWindow *RateLimitSliceWindow

	//statisticSliceWindow *metric.SliceWindow
	// 远程配额信息
	remoteQuota atomic.Value
	// 离线分配时的配额数
	// 对于local模式，tokenLocal = tokenTotal
	// 对于global模式，tokenLocal = tokenTotal/instCount
	tokenLocal int64
}

//CAS获取远程配额桶
func (t *TokenBucket) getRemoteQuota() *remoteQuotaInfo {
	var rQuotaInfo *remoteQuotaInfo
	for {
		rQuotaInfo = t.remoteQuota.Load().(*remoteQuotaInfo)
		rQuotaInfo.occupyWriter()
		if atomic.LoadInt32(&rQuotaInfo.updating) == 0 {
			return rQuotaInfo
		}
		rQuotaInfo.releaseWriter()
	}
}

// 令牌桶序列
type TokenBuckets []*TokenBucket

// 数组长度
func (tbs TokenBuckets) Len() int {
	return len(tbs)
}

// 比较数组成员大小
func (tbs TokenBuckets) Less(i, j int) bool {
	// 逆序
	return tbs[i].validDuration > tbs[j].validDuration
}

// 交换数组成员
func (tbs TokenBuckets) Swap(i, j int) {
	tbs[i], tbs[j] = tbs[j], tbs[i]
}

const metricTypeQps = "metricQps"

//计算每个实例分配的配额数
func getTokenPerInst(mode namingpb.Rule_Type, total int64, instCount int32) int64 {
	tokenPerInst := total
	if mode == namingpb.Rule_GLOBAL && instCount > 0 {
		if tokenPerInst = total / int64(instCount); tokenPerInst <= 0 {
			tokenPerInst = 1
		}
	}
	return tokenPerInst
}

// 初始化令牌桶
func initTokenBuckets(mode namingpb.Rule_Type, amounts []*namingpb.Amount, instCount int32,
	window *RateLimitWindow) TokenBuckets {
	buckets := make(TokenBuckets, 0, len(amounts))
	windowStartTime := clock.GetClock().Now().UnixNano()
	for _, amount := range amounts {
		goDuration, _ := pb.ConvertDuration(amount.GetValidDuration())
		total := int64(amount.GetMaxAmount().GetValue())
		tokenPerInst := getTokenPerInst(mode, total, instCount)
		bucket := &TokenBucket{
			validDuration: model.ToMilliSeconds(goDuration),
			tokenTotal:    total,
			tokenLocal:    tokenPerInst,
			sliceWindow: NewRateLimitSliceWindow(metricTypeQps, 1, goDuration, 1, windowStartTime,
				window),
		}
		bucket.remoteQuota.Store(&remoteQuotaInfo{})
		buckets = append(buckets, bucket)
	}
	if len(buckets) > 1 {
		sort.Sort(buckets)
	}
	return buckets
}
