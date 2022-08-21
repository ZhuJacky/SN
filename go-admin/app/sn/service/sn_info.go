package service

import (
	"errors"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/sn/models"
	"go-admin/app/sn/service/dto"
	"go-admin/common/actions"
	cDto "go-admin/common/dto"
)

type SNInfo struct {
	service.Service
}

// GetPage 获取BatchInfo列表
func (e *SNInfo) GetPage(c *dto.BatchInfoPageReq, list *[]models.BatchInfo, p *actions.DataPermission, count *int64) error {
	var err error
	var data models.BatchInfo

	// err = e.Orm.Model(&data).
	// 	Scopes(
	// 		cDto.MakeCondition(c.GetNeedSearch()),
	// 		cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
	// 	).
	// 	Find(list).Limit(-1).Offset(-1).
	// 	Count(count).Error
	err = e.Orm.Debug().Preload("Product").
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
			actions.Permission(data.TableName(), p),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error
	if err != nil {
		e.Log.Errorf("db error:%s \r", err)
		return err
	}
	return nil
}

// Get 获取SysPost对象
func (e *SNInfo) Get(d *dto.SNInfoGetReq, model *models.SNInfo) error {
	var err error
	var data models.SNInfo

	db := e.Orm.Model(&data).
		First(model, d.GetId())
	err = db.Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		err = errors.New("查看对象不存在或无权查看")
		e.Log.Errorf("db error:%s", err)
		return err
	}
	if db.Error != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	return nil
}

// GetPage 获取SNInfo列表
func (e *SNInfo) GetSNInfoList(c *dto.SNInfoPageReq, list *[]models.SNInfo, count *int64) error {
	var err error
	var data models.SNInfo

	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
		).
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error

	if err != nil {
		e.Log.Errorf("db error:%s \r", err)
		return err
	}
	return nil
}

// UpdateSNInfoStatus 修改SN状态
func (e *SNInfo) UpdateSNInfoStatus(c *dto.SNInfoUpdateReq) error {
	var err error
	var model = models.SNInfo{}
	e.Orm.First(&model, c.GetId())
	model.Status = c.Status
	model.SNId = c.SNId
	e.Log.Info("%v", &model)
	db := e.Orm.Save(&model)

	if db.Error != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	return nil
}

// Update 修改SysPost对象
func (e *SNInfo) Update(c *dto.BatchInfoUpdateReq) error {
	var err error
	var model = models.BatchInfo{}
	e.Orm.First(&model, c.GetId())
	//err = e.GenerateUpdateID(&model, c)
	//if err != nil {
	//	return err
	//}
	c.Generate(&model)
	e.Log.Info("%v", &model)
	db := e.Orm.Save(&model)
	if db.Error != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	return nil
}

// Remove 删除SysPost
func (e *SNInfo) Remove(d *dto.BatchInfoDeleteReq) error {
	var err error
	var data models.BatchInfo

	db := e.Orm.Model(&data).Delete(&data, d.GetId())
	if db.Error != nil {
		err = db.Error
		e.Log.Errorf("Delete error: %s", err)
		return err
	}
	if db.RowsAffected == 0 {
		err = errors.New("无权删除该数据")
		return err
	}
	return nil
}

// GetPage 获取装箱列表信息
func (e *SNInfo) GetBoxInfoList(c *dto.BoxInfoPageReq, list *[]models.SNBoxInfo, count *int64) error {
	var err error
	var data models.SNBoxInfo

	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
		).Order("box_id desc").
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error

	if err != nil {
		e.Log.Errorf("db error:%s \r", err)
		return err
	}
	return nil
}

func (e *SNInfo) UpdateBoxSum(c *dto.BoxInfoUpdateReq) error {
	var err error
	var model = models.SNBoxInfo{}
	e.Orm.First(&model, c.GetId())
	model.BoxSum = c.BoxSum
	model.BoxId = c.BoxId
	e.Log.Info("%v", &model)
	db := e.Orm.Save(&model)

	if db.Error != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	if db.RowsAffected == 0 {
		return errors.New("无权更新该数据")
	}
	return nil
}

// GetPage 获取装箱列表信息
func (e *SNInfo) GetRelationBoxInfoList(c *dto.BoxRelationInfoPageReq, list *[]models.SNBoxRelation, count *int64) error {
	var err error
	var data models.SNBoxRelation

	err = e.Orm.Model(&data).
		Scopes(
			cDto.MakeCondition(c.GetNeedSearch()),
			cDto.Paginate(c.GetPageSize(), c.GetPageIndex()),
		).Order("box_relation_id desc").
		Find(list).Limit(-1).Offset(-1).
		Count(count).Error

	if err != nil {
		e.Log.Errorf("db error:%s \r", err)
		return err
	}
	return nil
}
