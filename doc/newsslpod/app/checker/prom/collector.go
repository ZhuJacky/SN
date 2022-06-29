// Package prom provides ...
package prom

import "github.com/prometheus/client_golang/prometheus"

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

// PromFastDetection fast detection
var PromFastDetection = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "fast_detection",
		Help: "快速检测计数",
	},
	[]string{"status"},
)

// PromFastDetectionByError by err
func PromFastDetectionByError(err error) {
	PromFastDetection.WithLabelValues("total").Inc()
	if err != nil {
		PromFastDetection.WithLabelValues("failed").Inc()
	}
}

// PromFullDetection full detection
var PromFullDetection = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "full_detection",
		Help: "全量检测计数",
	},
	[]string{"status"},
)

// PromFullDetectionByError by err
func PromFullDetectionByError(err error) {
	PromFullDetection.WithLabelValues("total").Inc()
	if err != nil {
		PromFullDetection.WithLabelValues("failed").Inc()
	}
}

// PromMySSLOpenAPIErr myssl open api
var PromMySSLOpenAPIErr = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "myssl_openapi_err",
		Help: "调用myssl openapi计数",
	},
)
