// Package services provides ...
package services

import (
	"errors"
	"net/http"
	"time"

	"mysslee_qcloud/app/backend/aggregation"
	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/app/backend/prom"
	"mysslee_qcloud/model"
	"mysslee_qcloud/qcloud"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// register api to qcloud
func init() {
	qcloud.Register("DescribeProfile", HandleAccountProfile)
}

// HandleProfile get account info
// @Summary get account profile
// @Description account profile
// @Tags 账户信息
// @Accept json
// @Produce json
// @Param action query string true "DescribeProfile"
// @Param serviceType query string true "sslpod"
// @Success 200 {object} model.AccountShow
// @Router /profile [get]
func HandleAccountProfile(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/profile").Inc()

	a := req.GetValue("account").(*model.Account)
	// 修正套餐信息
	plan, err := db.GetCalculatedLimit(a.Uin)
	if err != nil {
		logrus.Error("HandleAccountProfile.GetCalculatedLimit: ", err)
		return nil, qcloud.ErrFailedOperation
	}

	show := model.AccountForShow(a) // 返回用户数据
	show.Plan = plan

	return &qcloud.ResponseBack{
		Response: gin.H{
			"Data":      show,
			"RequestId": req.GetString("RequestId"),
		},
	}, ""
}

// HandleAccountPlan get account plan
// @Summary get account plan or refresh
// @Description account plan
// @Tags 内部调用
// @Accept mpfd
// @Produce json
// @Param uin path int true "账户ID"
// @Success 200 {string} string "ok"
// @Router /plan/{uin} [get]
func HandleAccountPlan(c *gin.Context) {
	uin := c.Param("uin")
	plan, err := db.GetCalculatedLimit(uin)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, plan)
}

// GetAccount get account if not will create
func GetAccount(uin, ip string) (*model.Account, error) {
	if uin == "" {
		return nil, errors.New("not found uin")
	}

	// return &model.Account{
	// 	Id:  1,
	// 	Uin: uin,
	// }, nil

	// get account
	account, err := db.GetAccountByUin(uin)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			logrus.Error("GetAccount.GetAccount: ", err)
			return nil, err
		}
		// resp, err := qcloud.GetAccountInfoFromQCloud(uin)
		// if err != nil {
		// 	return nil, err
		// }
		// create account
		account = &model.Account{
			Uin:       uin,
			Name:      "",
			Status:    model.StatusActivated,
			CreatedAt: time.Now(),
		}
		err = db.AddAccount(account)
		if err != nil {
			logrus.Error("CreateAccount.AddAccount ", err)
			return nil, err
		}
		// not default
		nInfo := &model.NoticeInfo{
			Uin:        account.Uin,
			NoticeType: model.NoticeAll,
		}
		err = db.AddOrUpNoticeInfo(nInfo)
		if err != nil {
			logrus.Error("CreateAccount.AddOrUpNoticeInfo ", err)
		}
	}
	if account.Aggregate > 0 {
		triggerAggregation(account.Id)
	}
	return account, nil
}

// 触发聚合
func triggerAggregation(aid int) {
	err := db.UpAccountFiled(&model.Account{Id: aid}, map[string]interface{}{
		"aggregate": 0,
	})
	if err != nil {
		logrus.Error("triggerAggregation.UpAccountFiled: ", err)
		return
	}
	agg := &aggregation.AggregateType{
		AccountId:  aid,
		FromDomain: false, // 聚合指定用户
	}
	aggregation.AggrHandler.SendAggregateRequest(agg)
}
