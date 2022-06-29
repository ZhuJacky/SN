// Package payment provides ...
package payment

import (
	"encoding/json"
	"time"

	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/app/backend/db/redis"
	"mysslee_qcloud/model"
	"mysslee_qcloud/qcloud"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

func init() {
	qcloud.Register("checkModify", HandleCheckModify)
	qcloud.Register("modifyResource", HandleModifyResource)

	qcloud.Register("queryFlow", HandleQueryFlow)

	qcloud.Register("isolateResource", HandleIsolateResource)
}

// 变配参数检查
func HandleCheckModify(req *qcloud.Request) (interface{}, string) {
	uin := req.GetString("uin")

	goods := modifyGoodsParams(req, uin)
	if goods == nil {
		return result{"status": 1}, ""
	}
	return result{"status": 0}, ""
}

// 变更发货
func HandleModifyResource(req *qcloud.Request) (interface{}, string) {
	// 防重入
	tranId := req.GetString("tranId")
	_, err := db.GetOrderByTranId(tranId)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, qcloud.ErrInternalError
		}
	} else {
		return result{}, ""
	}

	// get goods
	uin := req.GetString("uin")
	goods := validModifyGoodsParmas(req, uin)
	if goods == nil {
		return nil, qcloud.ErrFailedOperation
	}
	resource, err := db.GetResourceByResourceId(goods.ResourceId)
	if err != nil {
		logrus.Error("HandleModifyResource.GetResourceByResourceId: ", err)
		if gorm.IsRecordNotFoundError(err) {
			return nil, qcloud.ErrResourceNotFound
		}
		return nil, qcloud.ErrInternalError
	}
	detail, err := resource.GetGoodsDetail()
	if err != nil {
		logrus.Error("HandleModifyResource.GetGoodsDetail: ", err)
		return nil, qcloud.ErrInternalError
	}
	// 获取新配置参数
	detail.Pid = goods.Pid

	data, err := json.Marshal(detail)
	if err != nil {
		logrus.Error("HandleModifyResource.Marshal: ", err)
		return nil, qcloud.ErrInternalError
	}
	resource.GoodsDetail = json.RawMessage(data)
	flowId, err := db.ModifyResource(resource, tranId)
	if err != nil {
		logrus.Error("HandleModifyResource.ModifyResource: ", err)
		return nil, qcloud.ErrInternalError
	}
	// 更新套餐
	redis.DelPlanLimit(resource.Uin)
	return result{"flowId": flowId}, ""
}

// 查询发货状态
func HandleQueryFlow(req *qcloud.Request) (interface{}, string) {
	flowId := req.GetInt("flowId")
	if flowId < 1 {
		return nil, qcloud.ErrInvalidParameterValue
	}
	// 查询订单
	_, err := db.GetOrderByFlowId(flowId)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return result{"status": 1}, ""
		}
		return nil, qcloud.ErrInternalError
	}
	// 更新套餐
	uin := req.GetString("uin")
	err = redis.DelPlanLimit(uin)
	if err != nil {
		logrus.Error("HandleQueryFlow.DelPlanLimit: ", err)
		return result{"status": 2}, ""
	}
	return result{"status": 0}, ""
}

// 隔离资源
func HandleIsolateResource(req *qcloud.Request) (interface{}, string) {
	resourceId := req.GetString("resourceId")
	resource, err := db.GetResourceByResourceId(resourceId)
	if err != nil {
		logrus.Error("HandleIsolateResource.GetResourceByResourceId: ", err)
		if gorm.IsRecordNotFoundError(err) {
			return nil, qcloud.ErrResourceNotFound
		}
		return nil, qcloud.ErrInternalError
	}
	uin := req.GetString("uin")
	if uin != resource.Uin {
		return nil, qcloud.ErrFailedOperation
	}

	renewFlag := req.GetInt("renewFlag")
	newDeadline := req.GetString("newDeadline")
	t, err := time.Parse(model.TIME_FORMAT, newDeadline)
	if err != nil || t.Before(resource.ExpireTime) {
		return nil, qcloud.ErrInvalidParameterValue
	}
	billingIsolateType := req.GetString("billingIsolateType")
	if billingIsolateType != "" &&
		billingIsolateType != model.OrderStatusRefund {
		return nil, qcloud.ErrInvalidParameterValue
	}

	resource.RenewFlag = renewFlag
	resource.ExpireTime = t
	err = db.IsolateResource(resource, billingIsolateType)
	if err != nil {
		logrus.Error("HandleIsolateResource.IsolateResource: ", err)
		return nil, qcloud.ErrInternalError
	}
	err = redis.DelPlanLimit(uin)
	if err != nil {
		return nil, qcloud.ErrInternalError
	}
	return result{}, ""
}
