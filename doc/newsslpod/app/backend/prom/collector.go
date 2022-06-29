package prom

import "github.com/prometheus/client_golang/prometheus"

// 用户数量
var PromAccountCount = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "account_count",
		Help: "用户数量",
	})

// 监控站点数
var PromMonitorSiteCount = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "site_count",
		Help: "监控站点数",
	},
)

// PromTaskDispatch task dispatch
var PromTaskDispatch = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "task_dispatch",
		Help: "调度任务的数量",
	},
	[]string{"node", "status"},
)

// PromBoughtPlan user plan
var PromBoughtPlan = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "user_plan",
		Help: "套餐购买数",
	},
	[]string{"plan"},
)

// PromDBError database status
var PromDBError = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "db_err",
		Help: "ping数据库错误",
	},
)

// PromRedisError redis err
var PromRedisError = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "redis_err",
		Help: "ping redis数据库",
	},
)

// PromApiRequest api request
var PromApiRequest = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "api_request",
		Help: "api 请求数",
	},
	[]string{"path"},
)

// PromRealtimePlan 实时购买
var PromRealtimePlan = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "realtime_plan",
		Help: "实时购买统计",
	},
	[]string{"uin", "plan", "timeSpan"},
)
