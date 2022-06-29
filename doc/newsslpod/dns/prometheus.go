// Package dns provides ...
package dns

import "github.com/prometheus/client_golang/prometheus"

// prometheus monitor
var PromDNS = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "dns_resolve_result",
		Help: "dns resolve status",
	},
	// 函数(sync_test, test)，完成情况(yes, no)
	[]string{"status"},
)

var PromDNSDown = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "dns_down",
		Help: "dns 挂掉次数",
	},
	[]string{"target"},
)

// DNS状态
const (
	DNSStateUP   float64 = 1
	DNSStateDown float64 = 0
)

// PromDNSState dns状态计数
var PromDNSState = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "dns_state",
		Help: "DNS 状态",
	},
	[]string{"target"},
)

// 计数器
func PromCount(err error) {
	PromDNS.WithLabelValues("total").Inc()
	if err != nil {
		PromDNS.WithLabelValues("failed").Inc()
		return
	}
	PromDNS.WithLabelValues("success").Inc()
}

// dns 在线个数
var PromDNSOnline = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "active_dns_server",
		Help: "dns 在线活动个数",
	})
