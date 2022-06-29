// Package payment provides ...
package payment

import (
	"encoding/json"
	"fmt"
	"time"

	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/app/backend/db/redis"
	"mysslee_qcloud/app/backend/prom"
	"mysslee_qcloud/model"
	"mysslee_qcloud/qcloud"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// 1. status 错误代码需要到计费前台配置
// 2. 同时购买套餐，发货问题？

func init() {
	qcloud.Register("checkCreate", HandleCheckCreate)
	qcloud.Register("createResource", HandleCreateResource)

	qcloud.Register("checkRenew", HandleCheckRenew)
	qcloud.Register("renewResource", HandleRenewResource)
	qcloud.Register("setRenewFlag", HandleSetRenewFlag)

	qcloud.Register("getAllAppIds", HandleGetAllAppIds)
	qcloud.Register("getUserResource", HandleGetUserResource)
	qcloud.Register("queryResources", HandleQueryResources)
	qcloud.Register("destroyResource", HandleDestroyResource)
	qcloud.Register("queryDeadlineList", HandleQueryDeadlineList)
}

// 检查参数返回
type result map[string]interface{}

// 新购参数检查
// func HandleCheckCreate(req *qcloud.Request) (interface{}, string) {
// 	uin := req.GetString("uin")
// 	// 检查相关参数
// 	_, errStr := checkGoodsParams(req, uin)
// 	if errStr != "" {
// 		logrus.Error("HandleCheckCreate.checkGoodsParams: ", errStr)
// 		return result{"status": 1}, ""
// 	}
// 	// appid
// 	appid := req.GetInt("appId")
// 	if !checkAppId(appid) {
// 		return result{"status": 1}, ""
// 	}
// 	_, err := db.GetResourceBoughtByUin(uin)
// 	if err != nil && !gorm.IsRecordNotFoundError(err) {
// 		logrus.Error("HandleCheckCreate.GetResourceBoughtByUin: ", err)
// 		return nil, qcloud.ErrInternalError
// 	} else if err == nil {
// 		// 已经购买，报错
// 		return result{"status": 1002132}, ""
// 	}
//
// 	return result{"status": 0}, ""
// }

// 新购参数检查
func HandleCheckCreate(req *qcloud.Request) (interface{}, string) {
	prom.PromApiRequest.WithLabelValues("checkCreate").Inc()

	uin := req.GetString("uin")
	// 检查相关参数
	goods, errStr := checkGoodsParams(req, uin)
	if errStr != "" {
		logrus.Error("HandleCheckCreate.checkGoodsParams: ", errStr)
		return result{"status": 1}, ""
	}
	// appid
	appid := req.GetInt("appId")
	if !checkAppId(appid) {
		return result{"status": 1}, ""
	}
	resource, err := db.GetResourceBoughtByUin(uin)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		logrus.Error("HandleCheckCreate.GetResourceBoughtByUin: ", err)
		return nil, qcloud.ErrInternalError
	} else if err == nil {
		detail, err := resource.GetGoodsDetail()
		if err != nil {
			logrus.Error("HandleCheckCreate.GetGoodsDetail: ", err)
			return nil, qcloud.ErrInternalError
		}
		// 已经购买, 但与现有套餐相同pid：过
		if goods.Pid != detail.Pid {
			return result{"status": 1002132}, ""
		}
	}

	return result{"status": 0}, ""
}

// 新购发货
func HandleCreateResource(req *qcloud.Request) (interface{}, string) {
	prom.PromApiRequest.WithLabelValues("createResource").Inc()

	// 防重入
	tranId := req.GetString("tranId")
	order, err := db.GetOrderByTranId(tranId)
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, qcloud.ErrInternalError
		}
	} else {
		return result{
			"resourceIds": []string{order.ResourceId},
		}, ""
	}

	// 参数是否和以前check一样
	uin := req.GetString("uin")
	goods, errStr := checkGoodsParams(req, uin)
	if errStr != "" {
		return nil, qcloud.ErrFailedOperation
	}
	// 获取现有资源，如果购买的是和现有资源是一致的，直接走续费流程
	resource, err := db.GetResourceBoughtByUin(uin)
	if err == nil {
		resource.ExpireTime = parseAddTime(resource.ExpireTime, goods.TimeUnit, goods.TimeSpan)
		resource.RenewFlag = goods.AutoRenewFlag
		flowId, err := db.RenewResource(resource, tranId)
		if err != nil {
			logrus.Error("HandleRenewResource.RenewResource: ", err)
			return nil, qcloud.ErrInternalError
		}
		// 刷新套餐缓存
		redis.DelPlanLimit(resource.Uin)

		prom.PromRealtimePlan.WithLabelValues(uin, fmt.Sprint(goods.Pid), fmt.Sprint(goods.TimeSpan))
		return result{
			"flowId":      flowId,
			"resourceIds": []string{resource.ResourceId},
		}, ""
	} else if !gorm.IsRecordNotFoundError(err) {
		logrus.Error("HandleCreateResource.GetResourceBoughtByUin: ", err)
		return nil, qcloud.ErrInternalError
	}

	detail := req.GetStringMap("goodsDetail")
	data, err := json.Marshal(detail)
	if err != nil {
		return nil, qcloud.ErrInternalError
	}
	rId, err := redis.UniqueResrouceId()
	if err != nil {
		return nil, qcloud.ErrInternalError
	}
	// 创建资源
	resource = &model.Resource{
		ResourceId:        rId,
		Uin:               uin,
		AppId:             req.GetInt("appId"),
		ProjectId:         req.GetInt("projectId"),
		RenewFlag:         goods.AutoRenewFlag,
		Region:            req.GetInt("region"),
		ZoneId:            req.GetInt("zoneId"),
		Status:            1,
		PayMode:           req.GetInt("payMode"),
		IsolatedTimestamp: model.TimeZeroAt,
		ExpireTime:        parseAddTime(time.Now(), goods.TimeUnit, goods.TimeSpan),
		GoodsDetail:       json.RawMessage(data),
	}
	flowId, err := db.AddResource(resource, tranId)
	if err != nil {
		logrus.Error("HandleCreateResource.AddResource: ", err)
		return nil, qcloud.ErrInternalError
	}
	//  更新套餐
	redis.DelPlanLimit(resource.Uin)

	prom.PromRealtimePlan.WithLabelValues(uin, fmt.Sprint(goods.Pid), fmt.Sprint(goods.TimeSpan))
	return result{
		"flowId":      flowId,
		"resourceIds": []string{rId},
	}, ""
}

// 续费参数检查
func HandleCheckRenew(req *qcloud.Request) (interface{}, string) {
	prom.PromApiRequest.WithLabelValues("checkRenew").Inc()

	// 资源是否属于用户
	uin := req.GetString("uin")

	_, errStr := renewGoodsParams(req, uin)
	if errStr != "" {
		return result{"status": 1}, ""
	}
	return result{"status": 0}, ""
}

// 续费发货
func HandleRenewResource(req *qcloud.Request) (interface{}, string) {
	prom.PromApiRequest.WithLabelValues("renewResource").Inc()

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

	uin := req.GetString("uin")

	goods, errStr := renewGoodsParams(req, uin)
	if errStr != "" {
		return nil, qcloud.ErrFailedOperation
	}

	resource := &model.Resource{
		ResourceId: goods.ResourceId,
		ExpireTime: parseAddTime(goods.ExpireTime, goods.TimeUnit, goods.TimeSpan),
		RenewFlag:  goods.AutoRenewFlag,
	}
	flowId, err := db.RenewResource(resource, tranId)
	if err != nil {
		logrus.Error("HandleRenewResource.RenewResource: ", err)
		return nil, qcloud.ErrInternalError
	}
	// 新套餐
	redis.DelPlanLimit(uin)

	prom.PromRealtimePlan.WithLabelValues(uin, fmt.Sprint(goods.Pid), fmt.Sprint(goods.TimeSpan))
	return result{"flowId": flowId}, ""
}

// 设置自动续费
func HandleSetRenewFlag(req *qcloud.Request) (interface{}, string) {
	prom.PromApiRequest.WithLabelValues("setRenewFlag").Inc()

	resourceIds := req.GetStringSlice("resourceIds")
	resources, err := db.GetResourcesByResourceIds(resourceIds)
	if err != nil {
		logrus.Error("HandleSetRenewFlag.GetResourceByResourceId: ", err)
		if gorm.IsRecordNotFoundError(err) {
			return nil, qcloud.ErrResourceNotFound
		}
		return nil, qcloud.ErrInternalError
	}
	autoRenewFlag := req.GetInt("autoRenewFlag")
	if autoRenewFlag != 1 && autoRenewFlag != 2 {
		return nil, qcloud.ErrInvalidParameterValue
	}
	uin := req.GetString("uin")
	for _, v := range resources {
		if uin != v.Uin {
			return nil, qcloud.ErrFailedOperation
		}
		detail, err := v.GetGoodsDetail()
		if err != nil {
			logrus.Error("HandleSetRenewFlag.GetGoodsDetail: ", err)
			return nil, qcloud.ErrInternalError
		}
		detail.AutoRenewFlag = autoRenewFlag
		v.RenewFlag = autoRenewFlag

		data, err := json.Marshal(detail)
		if err != nil {
			logrus.Error("HandleSetRenewFlag.Marshal: ", err)
			return nil, qcloud.ErrInternalError
		}
		v.GoodsDetail = json.RawMessage(data)
		err = db.UpdateResource(v)
		if err != nil {
			logrus.Error("HandleSetRenewFlag.UpdateResource: ", err)
			return nil, qcloud.ErrInternalError
		}
	}

	return result{}, ""
}

// 拉取拥有资源的所有用户appid
func HandleGetAllAppIds(req *qcloud.Request) (interface{}, string) {
	prom.PromApiRequest.WithLabelValues("getAllAppIds").Inc()

	pageNo := req.GetInt("pageNo")
	if pageNo < 0 {
		return nil, qcloud.ErrInvalidParameterValue
	}
	pageSize := req.GetInt("pageSize")
	if pageSize < 1 {
		return nil, qcloud.ErrInvalidParameterValue
	}

	appIds, total, err := db.GetResourcesAllAppIds(pageNo*pageSize, pageSize)
	if err != nil {
		logrus.Error("HandleGetAllAppIds.GetResourcesAllAppIds: ", err)
		return nil, qcloud.ErrInternalError
	}
	return result{
		"total":  total,
		"appIds": appIds,
	}, ""
}

// 获取用户的所有资源
func HandleGetUserResource(req *qcloud.Request) (interface{}, string) {
	prom.PromApiRequest.WithLabelValues("getUserResource").Inc()

	pageNo := req.GetInt("pageNo")
	if pageNo < 0 {
		return nil, qcloud.ErrInvalidParameterValue
	}
	pageSize := req.GetInt("pageSize")
	if pageSize < 1 {
		return nil, qcloud.ErrInvalidParameterValue
	}
	uin := req.GetString("uin")

	resource, err := db.GetResourceBoughtByUin(uin)
	if err != nil {
		logrus.Error("HandleGetUserResources.GetResourceBoughtByUin: ", err)
		if gorm.IsRecordNotFoundError(err) {
			return result{"resources": json.RawMessage("[]")}, ""
		}
		return nil, qcloud.ErrInternalError
	}

	m, err := showResource(resource)
	if err != nil {
		logrus.Error("HandleGetUserResources.showResource: ", err)
		return nil, qcloud.ErrInternalError
	}
	return result{
		"resources": []map[string]interface{}{m},
		"total":     1,
	}, ""
}

// 根据资源ID获取资源
func HandleQueryResources(req *qcloud.Request) (interface{}, string) {
	prom.PromApiRequest.WithLabelValues("queryResources").Inc()

	resourceIds := req.GetStringSlice("resourceIds")
	resources, err := db.GetResourcesByResourceIds(resourceIds)
	if err != nil {
		logrus.Error("HandleQueryResources.GetResourceByResourceId: ", err)
		if gorm.IsRecordNotFoundError(err) {
			return nil, qcloud.ErrResourceNotFound
		}
		return nil, qcloud.ErrInternalError
	}
	uin := req.GetString("uin")

	maps := make([]map[string]interface{}, len(resources))
	for i, v := range resources {
		if uin != v.Uin {
			return nil, qcloud.ErrFailedOperation
		}
		m, err := showResource(v)
		if err != nil {
			return nil, qcloud.ErrInternalError
		}
		maps[i] = m
	}
	return result{"resources": maps}, ""
}

// 回收资源
func HandleDestroyResource(req *qcloud.Request) (interface{}, string) {
	prom.PromApiRequest.WithLabelValues("destroyResource").Inc()

	resourceId := req.GetString("resourceId")
	resource, err := db.GetResourceByResourceId(resourceId)
	if err != nil {
		logrus.Error("HandleDestroyResource.GetResourceByResourceId: ", err)
		if gorm.IsRecordNotFoundError(err) {
			return nil, qcloud.ErrResourceNotFound
		}
		return nil, qcloud.ErrInternalError
	}
	uin := req.GetString("uin")
	if uin != resource.Uin {
		return nil, qcloud.ErrFailedOperation
	}
	err = db.DestroyResource(resourceId)
	if err != nil {
		logrus.Error("HandleDestroyResource.DestroyResource: ", err)
		return nil, qcloud.ErrInternalError
	}
	err = redis.DelPlanLimit(uin)
	if err != nil {
		return nil, qcloud.ErrInternalError
	}
	return result{}, ""
}

// 到期信息
func HandleQueryDeadlineList(req *qcloud.Request) (interface{}, string) {
	prom.PromApiRequest.WithLabelValues("queryDeadlineList").Inc()

	resource, err := db.GetResourceExpiringByUin(req.GetString("uin"))
	if err != nil {
		logrus.Error("HandleQueryDeadlineList.GetResourceExpiringByUin: ", err)
		return result{"instances": json.RawMessage("[]")}, ""
	}
	detail, err := resource.GetGoodsDetail()
	if err != nil {
		logrus.Error("HandleQueryDeadlineList.GetGoodsDetail: ", err)
		return result{}, qcloud.ErrInternalError
	}
	product, err := db.GetProductByPid(detail.Pid)
	if err != nil {
		logrus.Error("HandleQueryDeadlineList.GetProductByPid: ", err)
		return nil, qcloud.ErrInternalError
	}
	// search
	resourceIds := req.GetStringSlice("resourceIds")
	if len(resourceIds) > 0 {
		found := false
		for _, v := range resourceIds {
			if resource.ResourceId == v {
				found = true
				break
			}
		}
		if !found {
			return result{"instances": json.RawMessage("[]")}, ""
		}
	}
	deadlineStart := req.GetString("deadlineStart")
	if deadlineStart != "" {
		t, err := time.Parse(model.TIME_FORMAT, deadlineStart)
		if err != nil {
			return nil, qcloud.ErrInvalidParameterValue
		}
		if resource.ExpireTime.Before(t) {
			return result{"instances": json.RawMessage("[]")}, ""
		}
	}
	deadlineEnd := req.GetString("deadlineEnd")
	if deadlineEnd != "" {
		t, err := time.Parse(model.TIME_FORMAT, deadlineEnd)
		if err != nil {
			return nil, qcloud.ErrInvalidParameterValue
		}
		if resource.ExpireTime.After(t) {
			return result{"instances": json.RawMessage("[]")}, ""
		}
	}
	instance := make(map[string]interface{})
	instance["autoRenewFlag"] = detail.AutoRenewFlag
	instance["deadline"] = resource.ExpireTime.Format(model.TIME_FORMAT)
	instance["resourceId"] = resource.ResourceId
	instance["projectId"] = resource.ProjectId
	instance["regionId"] = resource.Region
	instance["zoneId"] = resource.ZoneId

	goods := make(map[string]interface{})
	goods["resourceId"] = resource.ResourceId
	goods["pid"] = detail.Pid
	goods[product.Description] = 1
	goods["curDeadline"] = resource.ExpireTime.Format(model.TIME_FORMAT)
	goods["productInfo"] = json.RawMessage(fmt.Sprintf(`[{"name":"%s","value":"%s"}]`, "套餐名", product.Name))
	instance["goodsDetail"] = goods
	return result{"instances": []map[string]interface{}{instance}, "totalCnt": 1}, ""
}

func showResource(r *model.Resource) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	m["resourceId"] = r.ResourceId
	m["uin"] = r.Uin
	m["appId"] = r.AppId
	m["projectId"] = r.ProjectId
	m["renewFlag"] = r.RenewFlag
	m["region"] = r.Region
	m["zoneId"] = r.ZoneId
	m["status"] = r.Status
	m["payMode"] = r.PayMode
	if r.IsolatedTimestamp.Equal(model.TimeZeroAt) {
		m["isolatedTimestamp"] = model.TIME_ZERO
	} else {
		m["isolatedTimestamp"] = r.IsolatedTimestamp.Format(model.TIME_FORMAT)
	}
	m["createTime"] = r.CreateTime.Format(model.TIME_FORMAT)
	if r.ExpireTime.Equal(model.TimeZeroAt) {
		m["expireTime"] = model.TIME_ZERO
	} else {
		m["expireTime"] = r.ExpireTime.Format(model.TIME_FORMAT)
	}
	goods, err := r.GetGoodsDetail()
	if err != nil {
		return nil, err
	}
	product, err := db.GetProductByPid(goods.Pid)
	if err != nil {
		return nil, err
	}
	goodsDetail := make(map[string]interface{})
	goodsDetail["subProductCode"] = product.Description
	goodsDetail["pid"] = product.Pid
	goodsDetail[product.Description] = 1
	type productInfo struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	goodsDetail["productInfo"] = productInfo{"套餐名", product.Name}
	m["goodsDetail"] = goodsDetail
	return m, nil
}
