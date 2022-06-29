// Package payment provides ...
package payment

import (
	"sync"
	"time"

	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/model"
	"mysslee_qcloud/qcloud"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

var cache sync.Map

func checkGoodsParams(req *qcloud.Request, uin string) (*model.GoodsDetail, string) {
	pid := req.GetInt("goodsDetail.pid")
	product, err := db.GetProductByPid(pid)
	if err != nil {
		logrus.Error("checkGoodsParams.GetProductByPid: ", err)
		return nil, qcloud.ErrProductNotFound
	}
	// check parameters 1-3 年
	timeSpan := req.GetInt("goodsDetail.timeSpan")
	if !checkTimeSpan(timeSpan) {
		return nil, qcloud.ErrInvalidParameterValue
	}
	// 单位 y
	timeUnit := req.GetString("goodsDetail.timeUnit")
	if !checkTimeUnit(timeUnit, product.TimeUnit) {
		return nil, qcloud.ErrInvalidParameterValue
	}
	// 数量只能购买一个
	goodsNum := req.GetInt("goodsDetail.goodsNum")
	if !checkGoodsNum(goodsNum) {
		return nil, qcloud.ErrInvalidParameterValue
	}
	// autoRenewFlag
	autoRenewFlag := req.GetInt("goodsDetail.autoRenewFlag")
	if autoRenewFlag != 0 && autoRenewFlag != 1 {
		return nil, qcloud.ErrInvalidParameterValue
	}
	// sslpod
	sslpod := req.GetInt("goodsDetail." + product.Description)
	if !checkSSLPod(sslpod) {
		return nil, qcloud.ErrInvalidParameterValue
	}
	goods := &model.GoodsDetail{
		Pid:           pid,
		TimeUnit:      timeUnit,
		TimeSpan:      timeSpan,
		GoodsNum:      goodsNum,
		AutoRenewFlag: autoRenewFlag,
	}
	setGoodsDetailSSLPod(goods, product.Description)

	return goods, ""
}

func validGoodsParams(req *qcloud.Request, uin string) *model.GoodsDetail {
	// get cache
	temp, ok := cache.Load(uin)
	if !ok {
		return nil
	}
	goods, ok := temp.(*model.GoodsDetail)
	if !ok {
		return nil
	}
	// pid
	if req.GetInt("goodsDetail.pid") != goods.Pid {
		return nil
	}
	// timeSpan
	if req.GetInt("goodsDetail.timeSpan") != goods.TimeSpan {
		return nil
	}
	// timeunit
	if req.GetString("goodsDetail.timeUnit") != goods.TimeUnit {
		return nil
	}
	// goods num
	if req.GetInt("goodsDetail.goodsNum") != goods.GoodsNum {
		return nil
	}
	// autoRenewFlag
	if req.GetInt("goodsDetail.autoRenewFlag") != goods.AutoRenewFlag {
		return nil
	}
	// sslpod
	if req.GetInt("goodsDetail."+getGoodsDetailSSLPodName(goods)) != 1 {
		return nil
	}
	return goods
}

type modifyGoodsDetail struct {
	model.GoodsDetail
	ResourceId  string
	CurDeadline string
}

func modifyGoodsParams(req *qcloud.Request, uin string) *modifyGoodsDetail {
	oldConfig := req.GetStringMap("goodsDetail.oldConfig")
	newConfig := req.GetStringMap("goodsDetail.newConfig")

	resourceId := req.GetString("resourceId")
	resource, err := db.GetResourceByResourceId(resourceId)
	if err != nil {
		logrus.Error("modifyGoodsParams.GetResourceByResourceId: ", err)
		if gorm.IsRecordNotFoundError(err) {
			return nil
		}
		return nil
	}
	detail, err := resource.GetGoodsDetail()
	if err != nil {
		logrus.Error("modifyGoodsParams.GetGoodsDetail: ", err)
		return nil
	}
	// 旧资源pid是否相同
	oldPid, _ := oldConfig["pid"].(float64)
	if int(oldPid) != detail.Pid {
		return nil
	}

	// 验证新pid
	newPid, _ := newConfig["pid"].(float64)
	product, err := db.GetProductByPid(int(newPid))
	if err != nil {
		logrus.Error("modifyGoodsParams.IsExistProduct: ", err)
		return nil
	}
	// check parameters 1-3 年
	timeSpan, _ := newConfig["timeSpan"].(float64)
	if !checkTimeSpan(int(timeSpan)) {
		return nil
	}
	// 单位 y
	timeUnit, _ := newConfig["timeUnit"].(string)
	if !checkTimeUnit(timeUnit, detail.TimeUnit) {
		return nil
	}
	// sslpod
	sslpod, _ := newConfig[product.Description].(float64)
	if !checkSSLPod(int(sslpod)) {
		return nil
	}
	// 数量只能为1
	goodsNum := req.GetInt("goodsDetail.goodsNum")
	if !checkGoodsNum(goodsNum) {
		return nil
	}
	// 到期时间是否相同
	deadline := req.GetString("goodsDetail.curDeadline")
	if !checkDeadline(deadline, resource.ExpireTime) {
		return nil
	}

	goods := &modifyGoodsDetail{}
	goods.Pid = int(newPid)
	goods.TimeUnit = timeUnit
	goods.TimeSpan = int(timeSpan)
	goods.GoodsNum = goodsNum
	setGoodsDetailSSLPod(&goods.GoodsDetail, product.Description)
	goods.ResourceId = resourceId
	goods.CurDeadline = deadline
	cache.Store(uin, goods)
	return goods
}

func validModifyGoodsParmas(req *qcloud.Request, uin string) *modifyGoodsDetail {
	// get cache
	temp, ok := cache.Load(uin)
	if !ok {
		return nil
	}
	goods, ok := temp.(*modifyGoodsDetail)
	if !ok {
		return nil
	}

	newConfig := req.GetStringMap("goodsDetail.newConfig")
	// pid
	newPid, _ := newConfig["pid"].(float64)
	if int(newPid) != goods.Pid {
		return nil
	}
	// timeSpan
	timeSpan, _ := newConfig["timeSpan"].(float64)
	if int(timeSpan) != goods.TimeSpan {
		return nil
	}
	// timeunit
	timeUnit, _ := newConfig["timeUnit"].(string)
	if timeUnit != goods.TimeUnit {
		return nil
	}
	// sslpod
	sslpod, _ := newConfig[getGoodsDetailSSLPodName(&goods.GoodsDetail)].(float64)
	if sslpod != 1 {
		return nil
	}
	// goods num
	if req.GetInt("goodsDetail.goodsNum") != goods.GoodsNum {
		return nil
	}
	// resource id
	if req.GetString("resourceId") != goods.ResourceId {
		return nil
	}
	if req.GetString("goodsDetail.curDeadline") != goods.CurDeadline {
		return nil
	}
	return goods
}

type renewGoodsDetail struct {
	model.GoodsDetail

	ResourceId  string
	CurDeadline string
	ExpireTime  time.Time
}

func renewGoodsParams(req *qcloud.Request, uin string) (*renewGoodsDetail, string) {
	resourceId := req.GetString("resourceId")
	resource, err := db.GetResourceByResourceId(resourceId)
	if err != nil {
		logrus.Error("renewGoodsParams.GetResourceByResourceId: ", err)
		if gorm.IsRecordNotFoundError(err) {
			return nil, qcloud.ErrResourceNotFound
		}
		return nil, qcloud.ErrInternalError
	}
	// 资源是否属于 uin
	if uin != resource.Uin {
		return nil, qcloud.ErrResourceNotFound
	}
	// 到期时间是否相同
	deadline := req.GetString("goodsDetail.curDeadline")
	if !checkDeadline(deadline, resource.ExpireTime) {
		return nil, qcloud.ErrFailedOperation
	}
	detail, err := resource.GetGoodsDetail()
	if err != nil {
		return nil, qcloud.ErrInternalError
	}

	// check parameters 1-3 年
	timeSpan := req.GetInt("goodsDetail.timeSpan")
	if !checkTimeSpan(timeSpan) {
		return nil, qcloud.ErrInvalidParameterValue
	}
	// 单位 y
	timeUnit := req.GetString("goodsDetail.timeUnit")
	if !checkTimeUnit(timeUnit, detail.TimeUnit) {
		return nil, qcloud.ErrInvalidParameterValue
	}
	// 数量只能购买一个
	goodsNum := req.GetInt("goodsDetail.goodsNum")
	if !checkGoodsNum(goodsNum) {
		return nil, qcloud.ErrInvalidParameterValue
	}
	// autoRenewFlag
	autoRenewFlag := req.GetInt("goodsDetail.autoRenewFlag")
	if autoRenewFlag != 0 && autoRenewFlag != 1 {
		return nil, qcloud.ErrInvalidParameterValue
	}
	// sslpod
	name := getGoodsDetailSSLPodName(detail)
	sslpod := req.GetInt("goodsDetail." + name)
	if !checkSSLPod(sslpod) {
		return nil, qcloud.ErrInvalidParameterValue
	}

	goods := &renewGoodsDetail{}
	goods.TimeUnit = timeUnit
	goods.TimeSpan = timeSpan
	goods.GoodsNum = goodsNum
	goods.AutoRenewFlag = autoRenewFlag
	setGoodsDetailSSLPod(&goods.GoodsDetail, name)
	goods.ResourceId = resourceId
	goods.CurDeadline = deadline
	goods.ExpireTime = resource.ExpireTime
	return goods, ""
}

func validRenewGoodsParams(req *qcloud.Request, uin string) *renewGoodsDetail {
	// get cache
	temp, ok := cache.Load(uin)
	if !ok {
		return nil
	}
	goods, ok := temp.(*renewGoodsDetail)
	if !ok {
		return nil
	}

	timeSpan := req.GetInt("goodsDetail.timeSpan")
	if goods.TimeSpan != timeSpan {
		return nil
	}
	timeUnit := req.GetString("goodsDetail.timeUnit")
	if goods.TimeUnit != timeUnit {
		return nil
	}
	goodsNum := req.GetInt("goodsDetail.goodsNum")
	if goods.GoodsNum != goodsNum {
		return nil
	}
	deadline := req.GetString("goodsDetail.curDeadline")
	if goods.CurDeadline != deadline {
		return nil
	}
	autoRenewFlag := req.GetInt("goodsDetail.autoRenewFlag")
	if goods.AutoRenewFlag != autoRenewFlag {
		return nil
	}
	// sslpod
	name := getGoodsDetailSSLPodName(&goods.GoodsDetail)
	sslpod := req.GetInt("goodsDetail." + name)
	if sslpod != 1 {
		return nil
	}
	return goods
}

func setGoodsDetailSSLPod(goods *model.GoodsDetail, name string) {
	if name == model.SSLPodV2 {
		goods.SSLPodV2 = 1
		return
	} else if name == model.SSLPodV3 {
		goods.SSLPodV3 = 1
		return
	}
	goods.SSLPodV1 = 1
}

func getGoodsDetailSSLPodName(goods *model.GoodsDetail) string {
	if goods.SSLPodV2 > 0 {
		return "sslpod_v2"
	} else if goods.SSLPodV3 > 0 {
		return "sslpod_v3"
	}
	return "sslpod_v1"
}
