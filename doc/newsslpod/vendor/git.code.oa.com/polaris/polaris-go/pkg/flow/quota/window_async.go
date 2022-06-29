package quota

import (
	"git.code.oa.com/polaris/polaris-go/pkg/clock"
	"git.code.oa.com/polaris/polaris-go/pkg/log"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	rlimit "git.code.oa.com/polaris/polaris-go/pkg/model/pb/metric"
	namingpb "git.code.oa.com/polaris/polaris-go/pkg/model/pb/v1"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/serverconnector"
	"math"
	"sync/atomic"
	"time"
)

// 异步发送 acquire
func (r *RateLimitWindow) doAsyncRemoteAcquire(connector serverconnector.AsyncRateLimitConnector) {
	if r.Rule.GetType() == namingpb.Rule_LOCAL || r.configMode == model.ConfigQuotaLocalMode {
		return
	}

	if r.GetStatus() == Expired || r.GetStatus() == Deleted {
		log.GetBaseLogger().Debugf("[RateLimitAcquire]doAsyncRemoteAcquire expire or delete return")
		return
	}

	if r.GetStatus() == RemoteAcquireFail {
		log.GetBaseLogger().Warnf("[RateLimitAcquire]doAsyncRemoteAcquire RemoteAcquireFail")
		return
	}
	notFinish := atomic.LoadInt32(&r.acquireNotFinish)
	lastAcTime := atomic.LoadInt64(&r.lastAcquireTime)
	timeNow := time.Now().UnixNano()
	if notFinish != 0 && timeNow-lastAcTime < int64(time.Millisecond*200) {
		return
	}
	if notFinish != 0 {
		if timeNow-lastAcTime > int64(time.Millisecond*200) {
			atomic.AddInt32(&r.continuousSlowRemoteAcquireCount, 1)
			log.GetBaseLogger().Infof("RateLimitWindow OnRemoteAcquireResponse slow acquire response "+
				"nowTime:%d lastAcTime:%d svcKey:%s quotaKey:%s", timeNow, lastAcTime, r.SvcKey.String(), r.quotaKey)
		}
	} else {
		atomic.StoreInt32(&r.continuousSlowRemoteAcquireCount, 0)
	}
	_, request := r.acquireRequest()
	atomic.StoreInt64(&r.lastAcquireTime, timeNow)
	atomic.StoreInt32(&r.acquireNotFinish, 1)
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		log.GetBaseLogger().Debugf("[RateLimitAcquire] doRemoteAcquire req:%s\n", request.String())
	}
	err := connector.AsyncAcquire(request, r)
	if nil != err {
		log.GetBaseLogger().Errorf(
			"fail to call RateLimitService.Acquire, service %s, quotaKey %s, error is %s",
			r.SvcKey, r.quotaKey, err)
		return
	}
}

func (r *RateLimitWindow) doAsyncRemoteAcquireOnlyReport(curTime int64, rpDur time.Duration, limited uint32) {
	if r.Rule.GetType() == namingpb.Rule_LOCAL || r.configMode == model.ConfigQuotaLocalMode {
		return
	}
	if r.GetStatus() == RemoteAcquireFail {
		return
	}
	timeNow := time.Now().UnixNano()
	request := r.assembleAcquireRequestOnlyWithReport(curTime, rpDur, limited)
	atomic.StoreInt64(&r.lastAcquireTime, timeNow)
	//atomic.StoreInt32(&r.acquireNotFinish, 1)
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		log.GetBaseLogger().Debugf("[RateLimitAcquire] doRemoteAcquire req:%s\n", request.String())
	}
	connector := r.WindowSet.flowAssistant.asyncRLimitConnector
	err := connector.AsyncAcquire(request, r)
	if nil != err {
		log.GetBaseLogger().Errorf(
			"fail to call RateLimitService.Acquire, service %s, quotaKey %s, error is %s",
			r.SvcKey, r.quotaKey, err)
		return
	}
}

// 异步处理 acquire 回包
func (r *RateLimitWindow) OnRemoteAcquireResponse(resp *rlimit.RateLimitResponse) {
	status := r.GetStatus()
	if status == Expired || status == Deleted {
		return
	}
	atomic.StoreInt32(&r.acquireNotFinish, 0)
	if log.GetBaseLogger().IsLevelEnabled(log.DebugLog) {
		log.GetBaseLogger().Debugf("[RateLimitAcquire]OnRemoteAcquireResponse: %s ", resp.String())
	}
	serverTime := resp.GetTimestamp().GetValue()
	nowTime := time.Now().UnixNano()/int64(time.Millisecond) + r.GetRemoteDiff()
	diff := math.Abs(float64(nowTime - serverTime))
	if diff >= float64(r.allocatingBucket.GetMaxRemoteWait()) {
		log.GetBaseLogger().Warnf("RateLimitWindow OnRemoteAcquireResponse ignore acquire response "+
			"nowTime:%d serverTime:%d svcKey:%s quotaKey:%s", nowTime, serverTime, r.SvcKey.String(), r.quotaKey)
		return
	}

	if atomic.LoadInt64(&r.lastRemoteDealQuotaTime) >= serverTime {
		return
	}
	atomic.StoreInt64(&r.lastRemoteDealQuotaTime, serverTime)

	if resp.GetCode().Value/1000 != 200 {
		r.SetStatus(RemoteInitFail)
	} else {
		//上报过程中使用的配额要进行扣除
		timeNowMill := model.ParseMilliSeconds(clock.GetClock().Now().UnixNano())
		usedQuota := r.allocatingBucket.GetQuotaUsed(timeNowMill)
		r.allocatingBucket.SetRemoteQuota(&RemoteQuotaResult{
			CurrentUsage: usedQuota,
			RemoteQuotas: resp.GetSumUseds(),
			ServerTime:   resp.GetTimestamp().GetValue(),
		})
		r.SetStatus(Acquired)
	}
}

/*
// 异步发送 MetricReport
func (r *RateLimitWindow) doAsyncMetricReport(connector serverconnector.AsyncRateLimitConnector,
	syncConnector serverconnector.RateLimitConnector)  {
	if r.Rule.GetType() == namingpb.Rule_LOCAL || r.configMode == model.ConfigQuotaLocalMode {
		return
	}
	if r.metricServerInitStatus != Initialized {
		r.doMetricReportInit(syncConnector)
	}
	timeNow := clock.GetClock().Now()
	if timeNow.UnixNano() - r.lastReportTime < int64(DefaultStatisticReportPeriod) {
		return
	}
	size := len(r.statisticsSlice)
	timeNowUnix := timeNow.Unix()
	if timeNow.UnixNano() - r.lastReportTime < int64(DefaultStatisticReportPeriod) {
		return
	}

	nowIdx := timeNowUnix % int64(size)
	var reportDataList []*ReportElements
	var i int64 = 0
	for i=0; i<int64(size); i++ {
		idx := nowIdx - i
		if idx < 0 {
			idx += int64(size)
		}
		rData := r.statisticsSlice[idx].GetReportData(timeNowUnix - i)
		reportDataList = append(reportDataList, rData)
	}

	request := r.reportRequest(timeNow, reportDataList)
	err := connector.AsyncReport(request, r)
	if err != nil {
		log.GetBaseLogger().Errorf(
			"fail to call RateLimitService.doAsyncMetricReport, service %s, quotaKey %s, error is %s",
			r.SvcKey, r.quotaKey, err)
		return
	}
	atomic.StoreInt64(&r.lastReportTime, timeNow.UnixNano())
}

*/

// 异步处理 MetricReport 回包
/*
func (r *RateLimitWindow) OnMetricReportResponse(resp *rlimit.MetricResponse) {
	status := r.GetStatus()
	if status == Expired || status == Deleted {
		return
	}
	if resp.GetCode().Value/1000 == 404 {
		atomic.StoreInt32(&r.metricServerInitStatus, Initializing)
	}
}

*/
