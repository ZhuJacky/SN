package db

import (
	"encoding/json"
	"fmt"
	"time"

	"mysslee_qcloud/app/backend/db/redis"
	"mysslee_qcloud/model"

	"github.com/jinzhu/gorm"
)

const (
	ResourceStatusNormal  = 1
	ResourceStatusIsolate = 2
	ResourceStatusDestroy = 3
)

var (
	mResource = model.Resource{}
	mOrder    = model.Order{}
)

// 获取限制
func GetCalculatedLimit(uin string) (plan *model.PlanInfo, err error) {
	plan = new(model.PlanInfo)
	// 获取redis缓存
	data, err := redis.GetPlanLimit(uin)
	if err == nil {
		err = json.Unmarshal(data, plan)
		return
	}

	// 查询有效资源
	resource, err := GetResourceByUin(uin)
	if err != nil {
		return
	}
	// 查询产品
	detail, err := resource.GetGoodsDetail()
	if err != nil {
		return
	}
	product := new(model.Product)
	err = gormDB.Where("pid=?", detail.Pid).First(product).Error
	if err != nil {
		return
	}
	pc := new(model.ProductContent)
	err = json.Unmarshal(product.Content, pc)
	if err != nil {
		return
	}

	plan.ProductContent = *pc
	plan.Name = product.Name
	plan.Pid = product.Pid
	if plan.Pid == 15958 {
		plan.ExpiredAt = "永久"
	} else {
		plan.ExpiredAt = resource.ExpireTime.Format("2006-01-02 15:04:05")
	}
	// 保存到redis
	data, err = json.Marshal(plan)
	if err != nil {
		return
	}
	// 取消到期约束，，由计费侧通知
	err = redis.SetPlanLimit(uin, data, 0)
	return
}

/// 获取用户购买的有效资源
func GetResourceBoughtByUin(uin string) (*model.Resource, error) {
	resource := new(model.Resource)
	// 查看有效付费资源
	err = gormDB.Where("uin=? AND expire_time>now() AND status=?", uin, ResourceStatusNormal).
		First(resource).Error
	return resource, err
}

// 获取用户到期资源
func GetResourceExpiringByUin(uin string) (*model.Resource, error) {
	resource := new(model.Resource)
	// 查看过期
	err = gormDB.Where("uin=? AND renew_flag=? AND status=?", uin, 1, ResourceStatusNormal).
		First(resource).Error
	return resource, err
}

// 获取用户现有资源，包括基础资源
func GetResourceByUin(uin string) (*model.Resource, error) {
	resource := new(model.Resource)
	// 查看有效付费资源
	err = gormDB.Where("uin=? AND expire_time>now() AND status=?", uin, ResourceStatusNormal).
		First(resource).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		// 无有效订单，查询基础资源
		resource.ResourceId = DEFAULT_PLAN_RESOURCE
		err = gormDB.First(resource).Error
		if err != nil {
			return nil, err
		}
	}

	return resource, nil
}

// 获取资源
func GetResourceByResourceId(rId string) (*model.Resource, error) {
	resource := new(model.Resource)
	err := gormDB.Where("resource_id=?", rId).First(resource).Error
	return resource, err
}

// 资源是否存在
func IsExistResource(resourceId string) bool {
	var count int
	err := gormDB.Model(mResource).Where("resource_id=?", resourceId).
		Count(&count).Error
	return err == nil && count > 0
}

// 添加资源
func AddResource(resource *model.Resource, tranId string) (flowId int, err error) {
	tx := gormDB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 创建资源
	err = tx.Create(resource).Error
	if err != nil {
		return
	}
	// 创建订单
	order := &model.Order{
		TranId:     tranId,
		ResourceId: resource.ResourceId,
		Status:     model.OrderStatusValid,
	}
	err = tx.Create(order).Error
	return order.Id, err
}

// 添加资源无flowId
func AddResourceWithoutFlowId(resource *model.Resource) error {
	return gormDB.Create(resource).Error
}

// 修改资源
func ModifyResource(resource *model.Resource, tranId string) (flowId int, err error) {
	tx := gormDB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = gormDB.Model(mResource).Where("resource_id=?", resource.ResourceId).
		Updates(map[string]interface{}{
			"expire_time":  resource.ExpireTime,
			"goods_detail": resource.GoodsDetail,
		}).Error
	if err != nil {
		return
	}
	// new order
	order := &model.Order{
		TranId:     tranId,
		ResourceId: resource.ResourceId,
		Status:     model.OrderStatusValid,
	}
	err = tx.Create(order).Error
	return order.Id, err
}

// 更新资源
func UpdateResource(resource *model.Resource) error {
	return gormDB.Model(mResource).Where("resource_id=?", resource.ResourceId).
		Updates(map[string]interface{}{
			"goods_detail": resource.GoodsDetail,
			"renew_flag":   resource.RenewFlag,
		}).Error
}

// 续费资源
func RenewResource(resource *model.Resource, tranId string) (flowId int, err error) {
	tx := gormDB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// renew 将订单状态置为更新
	err = tx.Model(mOrder).Where("resource_id=? AND status=?",
		resource.ResourceId, model.OrderStatusValid).
		Update("status", model.OrderStatusRenew).Error
	if err != nil {
		return
	}
	// update resource，更新资源信息
	err = tx.Model(mResource).Where("resource_id=?", resource.ResourceId).
		Updates(map[string]interface{}{
			"expire_time": resource.ExpireTime,
			"renew_flag":  resource.RenewFlag,
		}).Error
	// new order，创建新的订单
	order := &model.Order{
		TranId:     tranId,
		ResourceId: resource.ResourceId,
		Status:     model.OrderStatusValid,
	}
	err = tx.Create(order).Error
	return order.Id, err
}

// 销毁资源
func DestroyResource(rId string) (err error) {
	tx := gormDB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	// 销毁订单
	err = tx.Model(mOrder).Where("resource_id=?", rId).
		Update("status", model.OrderStatusDestroy).Error
	if err != nil {
		return
	}
	// 销毁资源
	return gormDB.Model(mResource).Where("resource_id=?", rId).
		Update("status", ResourceStatusDestroy).Error
}

// 隔离资源
func IsolateResource(resource *model.Resource, typ string) (err error) {
	tx := gormDB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if typ != "" {
		err = tx.Model(mOrder).Where("resource_id=?", resource.ResourceId).
			Update("status", typ).Error
		if err != nil {
			return
		}
	}

	return tx.Model(mResource).Where("resource_id=?", resource.ResourceId).
		Updates(map[string]interface{}{
			"renew_flag":         resource.RenewFlag,
			"expire_time":        resource.ExpireTime,
			"isolated_timestamp": time.Now(),
		}).Error
}

// 获取多个资源
func GetResourcesByResourceIds(ids []string) ([]*model.Resource, error) {
	var resources []*model.Resource
	for _, id := range ids {
		resource := model.Resource{}
		err := gormDB.Where("resource_id=?", id).First(&resource).Error
		if err != nil {
			return nil, err
		}
		resources = append(resources, &resource)
	}
	return resources, nil
}

// 设置自动自费flag
func SetResourcesAutoRenewFlag(resources []*model.Resource) error {
	for _, v := range resources {
		err := gormDB.Model(mResource).Where("resource_id=?", v.ResourceId).
			Updates(map[string]interface{}{
				"goods_detail": v.GoodsDetail,
				"renew_flag":   v.RenewFlag,
			}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// 获取拥有资源的所有用户appid
func GetResourcesAllAppIds(offset, limit int) ([]int, int, error) {
	var (
		appIds = make([]int, 0)
		total  int
	)

	err := gormDB.Select("COUNT(DISTINCT app_id)").Model(mResource).
		Joins("JOIN orders ON orders.resource_id=resources.resource_id").
		Where("orders.status=? AND resources.status=? AND expire_time>now()",
			model.OrderStatusValid, ResourceStatusNormal).Row().Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := gormDB.Select("DISTINCT app_id").Model(mResource).
		Joins("JOIN orders ON orders.resource_id=resources.resource_id").
		Where("orders.status=? AND resources.status=? AND expire_time>now()",
			model.OrderStatusValid, ResourceStatusNormal).
		Limit(limit).
		Offset(offset).
		Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var appId int
		err = rows.Scan(&appId)
		if err != nil {
			return nil, 0, err
		}
		appIds = append(appIds, appId)
	}

	return appIds, total, err
}

// 获取order信息
func GetOrderByTranId(tranId string) (*model.Order, error) {
	order := new(model.Order)
	err := gormDB.Where("tran_id=?", tranId).First(order).Error
	return order, err
}

// 获取order信息
func GetOrderByFlowId(flowId int) (*model.Order, error) {
	order := new(model.Order)
	err := gormDB.Where("id=?", flowId).First(order).Error
	return order, err
}

// 统计套餐订单数
func GetResourcePlanCount() (map[string]int, error) {
	var orders []*model.Order
	err := gormDB.Model(mOrder).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]int)
	// 获取有效订单
	for _, v := range orders {
		if v.Status == model.OrderStatusDestroy { // 销毁订单
			dur := v.UpdatedAt.Sub(v.CreatedAt).Hours()

			if dur < 20*24 { // 20天内视为无效订单
				continue
			}
		}
		// 获取资源
		resource := new(model.Resource)
		err = gormDB.Model(mResource).Select("goods_detail").
			Where("resource_id=?", v.ResourceId).
			First(resource).Error
		if err != nil {
			return nil, err
		}
		detail := new(model.GoodsDetail)
		err = json.Unmarshal(resource.GoodsDetail, detail)
		if err != nil {
			return nil, err
		}
		result[fmt.Sprint(detail.Pid)+":"+fmt.Sprint(detail.TimeSpan)] += 1
	}
	return result, nil
}
