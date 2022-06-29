// Package prom provides ...
package prom

import "github.com/prometheus/client_golang/prometheus"

// PromRedisError redis err
var PromRedisError = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "redis_err",
		Help: "ping redis数据库",
	},
)

// PromNoticeStatus 通知数量
var PromNoticeStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "notice_status",
		Help: "通知数量状态",
	},
	[]string{"status"},
)
