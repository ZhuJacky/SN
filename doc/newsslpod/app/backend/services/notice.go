// Package services provides ...
package services

import (
	"time"

	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/app/backend/prom"
	"mysslee_qcloud/model"
	"mysslee_qcloud/qcloud"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

func init() {
	qcloud.Register("DescribeNoticeInfo", HandleNoticeInfo)
	qcloud.Register("ModifyNoticeType", HandleModifyNoticeType)
}

// NoticeInfo 通知信息
type NoticeInfo struct {
	Id         int
	NoticeType int
	LimitInfos []*model.LimitInfo
}

// HandleNoticeInfo 获取通知信息
// @Summary 获取通知信息
// @Description 获取通知信息
// @Tags 通知告警
// @Accept mpfd
// @Produce json
// @Param action query string true "DescribeNoticeInfo"
// @Param serviceType query string true "sslpod"
// @Success 200 {object} services.NoticeInfo "msg.Data, limit_infos中的type可能为:limit_email,limit_wechat,limit_phone"
// @Router /notice/info [get]
func HandleNoticeInfo(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/notice/info").Inc()

	data, errCode := coreNoticeInfo(req)
	return &qcloud.ResponseBack{
		Response: gin.H{
			"RequestId": req.GetString("RequestId"),
			"Data":      data,
		},
	}, errCode
}

// coreNoticeInfo 内部调用(获取通知信息)
func coreNoticeInfo(req qcloud.RequestBack) (*NoticeInfo, string) {
	a := req.GetValue("account").(*model.Account)

	plan, err := db.GetCalculatedLimit(a.Uin)
	if err != nil {
		logrus.Error("CoreNoticeInfo.GetCalculatedLimit: ", err)
		return nil, qcloud.ErrInternalError
	}
	nInfo, err := db.GetNoticeInfo(a.Uin)
	if err != nil { // 修正数据
		logrus.Error("CoreNoticeInfo.GetNoticeInfo: ", err)
		if err != gorm.ErrRecordNotFound {
			return nil, qcloud.ErrInternalError
		}
		nInfo = &model.NoticeInfo{
			Uin:        a.Uin,
			NoticeType: model.NoticeAll,
		}
		err = db.AddOrUpNoticeInfo(nInfo)
		if err != nil {
			logrus.Error("CoreNoticeInfo.AddOrUpNoticeInfo: ", err)
			return nil, qcloud.ErrInternalError
		}
	}
	consume, err := db.GetLimitConsume(a.Uin, time.Now().Format("2006-01"))
	if err != nil {
		logrus.Error("coreNoticeInfo.GetLimitConsume: ", err)
		return nil, qcloud.ErrInternalError
	}

	return &NoticeInfo{
		Id:         nInfo.Id,
		NoticeType: nInfo.NoticeType,
		LimitInfos: []*model.LimitInfo{
			&model.LimitInfo{
				Type:  model.EmailLimitName,
				Total: plan.MaxAllowEmailWarnCount,
				Sent:  consume.Cost,
			},
		},
	}, ""
}

// HandleModifyNoticeType 更改通知类型
// @Summary 更改通知类型
// @Description 通知类型更改，用二进制进行表示: 001微信，010短信，100邮箱。对应二进制位，如果是1代表开启,0代表关闭
// @Tags 通知告警
// @Accept mpfd
// @Produce json
// @Param action query string true "ModifyNoticeType"
// @Param serviceType query string true "sslpod"
// @Param Type formData int true "通知开关的二进制: 1 微信|2 短信|4 邮件"
// @Success 200 {object} qcloud.ResponseBack
// @Router /notice/switch [put]
func HandleModifyNoticeType(req qcloud.RequestBack) (*qcloud.ResponseBack, string) {
	prom.PromApiRequest.WithLabelValues("/notice/switch").Inc()

	errCode := coreModifyNoticeType(req)
	return &qcloud.ResponseBack{
		Response: gin.H{
			"RequestId": req.GetString("RequestId"),
		},
	}, errCode
}

// coreModifyNoticeType 内部调用(更改通知类型)
func coreModifyNoticeType(req qcloud.RequestBack) string {
	a := req.GetValue("account").(*model.Account)
	typ := req.GetInt("Type")
	if !model.IsSupportedNoticeType(typ) {
		return qcloud.ErrInvalidNoticeType
	}
	nInfo, err := db.GetNoticeInfo(a.Uin)
	if err != nil {
		logrus.Error("CoreNoticeChangeType.GetNoticeInfo: ", err)
		return qcloud.ErrInternalError
	}
	nInfo.NoticeType = typ
	err = db.AddOrUpNoticeInfo(nInfo)
	if err != nil {
		logrus.Error("CoreNoticeChangeType.AddOrUpNoticeInfo: ", err)
		return qcloud.ErrInternalError
	}
	return ""
}
