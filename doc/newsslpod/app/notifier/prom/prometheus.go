package prom

import (
	"mysslee_qcloud/prom"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(PromRedisError)
	prometheus.MustRegister(PromNoticeStatus)

	prom.InitProm("mysslee_notifier")
}
