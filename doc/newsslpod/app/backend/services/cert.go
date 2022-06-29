// Package services provides ...
package services

import (
	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/app/backend/prom"
	"mysslee_qcloud/model"
	"mysslee_qcloud/qcloud"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func init() {
	qcloud.Register("DescribeCerts", HandleCertList)
}

// HandleCertList 证书列表
// @Summary 证书列表
// @Description 获取已经添加域名的证书
// @Tags SSL证书
// @Accept json
// @Produce json
// @Param action query string true "DescribeCerts"
// @Param serviceType query string true "sslpod"
// @Param Offset query int true "偏移量"
// @Param Limit query int true "数量"
// @Param Search query string false "搜索的类型，可选：domainId,commonName,hash" Enums(domainId,commonName,hash)
// @Param Target query string false "搜索类型的值"
// @Success 200 {array} model.CertInfoShow "msg.Data"
// @Router /certs [get]
func HandleCertList(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/certs").Inc()

	a := req.GetValue("account").(*model.Account)

	// page
	offset := req.GetInt("Offset")
	if offset < 0 {
		offset = 0
	}
	limit := req.GetInt("Limit")
	if limit < 0 {
		limit = 10
	}
	// search
	search := req.GetString("Search")
	target := req.GetString("Target")
	if search != "" {
		if !db.SupportedSearch[search] {
			search = ""
		}
	}
	// get data
	list, total, err := db.GetDomainCertList(a.Id, offset, limit, search, target)
	if err != nil {
		logrus.Error("HandleCertList.GetDomainsCert: ", err)
		return nil, qcloud.ErrInternalError
	}
	show := model.CertInfoForShow(list)
	return &qcloud.ResponseBack{
		Response: gin.H{
			"Data": gin.H{
				"List":  show,
				"Total": total,
			},
			"RequestId": req.GetString("RequestId"),
		},
	}, ""
}
