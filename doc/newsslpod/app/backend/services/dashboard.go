package services

import (
	"encoding/json"

	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/app/backend/prom"
	"mysslee_qcloud/model"
	"mysslee_qcloud/qcloud"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func init() {
	qcloud.Register("DescribeDashboard", HandleDashboard)
}

// HandleDashboard 图表dashboard
// @Summary dashboard图表
// @Description 获取图表dashboard数据
// @Tags 监控面板
// @Accept mpfd
// @Produce json
// @Param action query string true "DescribeDashboard"
// @Param serviceType query string true "sslpod"
// @Success 200 {string} string "json string"
// @Router /dashboard [get]
func HandleDashboard(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/dashboard").Inc()

	a := req.GetValue("account").(*model.Account)

	data, err := db.GetDashboardResultByAccountId(a.Id)
	if err != nil {
		logrus.Error("HandleDashboard.GetDashboardResultByAccountId: ", err)
		return nil, qcloud.ErrInternalError
	}

	jmsg := json.RawMessage(data)
	return &qcloud.ResponseBack{
		Response: gin.H{
			"Data":      jmsg,
			"RequestId": req.GetString("RequestId"),
		},
	}, ""
}
