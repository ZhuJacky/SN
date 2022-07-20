package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-admin-team/go-admin-core/sdk/service"
	"gorm.io/gorm"

	"go-admin/app/sn/models"
	"go-admin/app/sn/service/dto"
	"go-admin/common/actions"
	cDto "go-admin/common/dto"
)

type BatchInfo struct {
	service.Service
}

// GetPage 获取BatchInfo列表
func (e *BatchInfo) GetPage(c *dto.BatchInfoPageReq, list *[]models.BatchInfo, p *actions.DataPermission, count *int64) error {
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
func (e *BatchInfo) Get(d *dto.BatchInfoGetReq, model *models.BatchInfo) error {
	var err error
	var data models.BatchInfo

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

func ProductFilter(db *gorm.DB, productID int) *gorm.DB {
	return db.Where("product_id = ?", productID)
}

func MonthFilter(db *gorm.DB, month string) *gorm.DB {
	date, _ := time.Parse("2006-01-02", month+"-03")
	return db.Where("product_month = ?", date)
}

func (e *BatchInfo) GenerateInsertID(model *models.BatchInfo, s *dto.BatchInfoInsertReq) error {
	var list []models.BatchInfo
	date, _ := time.Parse("2006-01-02", s.ProductMonth+"-03")
	e.Orm.Unscoped().Where("product_id= ? AND DATE_FORMAT(product_month,'%Y-%m')= ?", s.ProductId, s.ProductMonth).Find(&list)
	model.ProductMonth = date
	var sum int = 0
	for _, batch := range list {
		sum = sum + batch.BatchNumber + batch.BatchExtra
	}
	year := date.Year()
	ycode := (year - 33) % 100
	month := date.Month()
	mcode := month + 22
	smin := fmt.Sprintf("%06d", sum+1)
	smax := fmt.Sprintf("%06d", sum+s.BatchNumber+s.BatchExtra)
	model.SNMax = strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + s.ProductCode + smax
	model.SNMin = strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + s.ProductCode + smin
	count := len(list)
	var cstr string = fmt.Sprintf("%03d", count+1)
	monthStr := fmt.Sprintf("%02d", int(month))
	model.BatchCode = strconv.Itoa(year) + monthStr + cstr
	model.External = s.External
	if model.External == 1 {
		model.SNMax = "(01)" + model.SNMax
		model.SNMin = "(01)" + model.SNMin
	}
	return nil
}

// Insert 创建SysPost对象
func (e *BatchInfo) Insert(c *dto.BatchInfoInsertReq) error {
	var err error
	var data models.BatchInfo
	_ = e.GenerateInsertID(&data, c)
	c.Generate(&data)
	err = e.Orm.Create(&data).Error
	if err != nil {
		e.Log.Errorf("db error:%s", err)
		return err
	}
	//批量添加SN信息
	e.InsertSNInfo(&data)
	return nil
}


//添加SN信息
func (e *BatchInfo) InsertSNInfo(batch *models.BatchInfo) error {
	var err error
	var number int = batch.BatchNumber + batch.BatchExtra
	for i:=1;i<number;i++{
		var data models.SNInfo
		data.BatchId = batch.BatchId
		data.BatchName = batch.BatchName
		data.BatchCode = batch.BatchCode
		data.WorkCode = batch.WorkCode
		data.ProductCode = batch.ProductCode
		data.UDI = batch.UDI
		data.ProductMonth = batch.ProductMonth
		date := data.ProductMonth

		year := date.Year()
		ycode := (year - 33) % 100
		month := date.Month()
		mcode := month + 22
		sn := fmt.Sprintf("%06d", i)
		data.SNCode  = strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) +data.ProductCode+ sn
		External := batch.External
		if External == 1 {
			data.SNCode = "(01)" + data.SNCode
		}

		data.Status = batch.Status

		data.CreateBy = batch.CreateBy
		data.UpdateBy = batch.UpdateBy

		err = e.Orm.Create(&data).Error
		if err != nil {
			e.Log.Errorf("db error:%s", err)
			return err
		}
	}

	return nil
}

// GetPage 获取SNInfo列表
func (e *BatchInfo) GetSNInfoList(c *dto.SNInfoPageReq, list *[]models.SNInfo, count *int64) error {
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

func (e *BatchInfo) GenerateUpdateID(model *models.BatchInfo, s *dto.BatchInfoUpdateReq) error {
	var list []models.BatchInfo
	date, _ := time.Parse("2006-01-02", s.ProductMonth+"-03")
	e.Orm.Unscoped().Where("product_id= ? AND DATE_FORMAT(product_month,'%Y-%m')= ?", s.ProductId, s.ProductMonth).Find(&list)
	model.ProductMonth = date
	var sum int = 0
	var isLast bool = true
	var count int = 1
	for _, batch := range list {

		if batch.BatchId < model.BatchId {
			sum = sum + batch.BatchNumber + batch.BatchExtra
			count++
		} else if batch.BatchId > model.BatchId {
			isLast = false
		}
	}
	if !isLast {
		if s.BatchNumber+s.BatchExtra != model.BatchNumber+model.BatchExtra {
			return errors.New("不是当月最后一批，不要改变数量，以免影响其他批次")
		}
	}
	year := date.Year()
	ycode := (year - 33) % 100
	month := date.Month()
	mcode := month + 22
	smin := fmt.Sprintf("%06d", sum+1)
	var numMax int
	if s.BatchNumber+s.BatchExtra > model.BatchNumber+model.BatchExtra {
		numMax = s.BatchNumber + s.BatchExtra
	} else {
		numMax = model.BatchNumber + model.BatchExtra
	}
	smax := fmt.Sprintf("%06d", sum+numMax)
	SNMax := strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + s.ProductCode + smax
	SNMin := strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + s.ProductCode + smin

	var cstr string = fmt.Sprintf("%03d", count+1)
	monthStr := fmt.Sprintf("%02d", int(month))
	if s.External == 1 {
		model.External = s.External
		model.SNMax = "(01)" + SNMax
		model.SNMin = "(01)" + SNMin
	} else {
		model.External = s.External
		model.SNMax = SNMax
		model.SNMin = SNMin
	}
	model.BatchCode = strconv.Itoa(year) + monthStr + cstr
	return nil
}

// Update 修改SysPost对象
func (e *BatchInfo) Update(c *dto.BatchInfoUpdateReq) error {
	var err error
	var model = models.BatchInfo{}
	e.Orm.First(&model, c.GetId())
	err = e.GenerateUpdateID(&model, c)
	if err != nil {
		return err
	}
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
func (e *BatchInfo) Remove(d *dto.BatchInfoDeleteReq) error {
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
