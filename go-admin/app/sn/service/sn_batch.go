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
	//手动填写的LOT号，不需要占用自动生成批号的批次
	e.Orm.Unscoped().Where("product_id= ? AND DATE_FORMAT(product_month,'%Y-%m')= ?", s.ProductId, s.ProductMonth).Find(&list)
	model.ProductMonth = date
	var autoSNSum int = 0
	var autoBatchCount int = 1
	for _, batch := range list {
		if batch.SNCodeRules == 0 {
			autoSNSum = autoSNSum + batch.BatchNumber + batch.BatchExtra
		}
		if batch.BatchCodeFormat == 0 {
			autoBatchCount++
		}
	}
	year := date.Year()
	ycode := (year - 33) % 100
	month := date.Month()
	mcode := month + 22
	smin := fmt.Sprintf("%06d", autoSNSum+1)
	smax := fmt.Sprintf("%06d", autoSNSum+s.BatchNumber+s.BatchExtra)
	model.SNMax = strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + s.ProductCode + smax
	model.SNMin = strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + s.ProductCode + smin

	var cstr string = fmt.Sprintf("%03d", autoBatchCount)
	monthStr := fmt.Sprintf("%02d", int(month))
	model.BatchCode = strconv.Itoa(year) + monthStr + cstr
	model.External = s.External
	model.SNFormat = s.SNFormat
	model.SNFormatInfo = s.SNFormatInfo
	/*
		if model.External == 1 {
			model.SNMax = "(01)" + model.SNMax
			model.SNMin = "(01)" + model.SNMin
		}*/

	//是否手动填写LOT号
	model.BatchCodeFormat = s.BatchCodeFormat
	e.Log.Info("aaaaa %v", s.BatchCodeFormatInfo)
	if model.BatchCodeFormat == 1 {
		model.BatchCode = s.BatchCodeFormatInfo
		model.BatchCodeFormatInfo = s.BatchCodeFormatInfo
	}

	//客户指定SN号
	model.SNCodeRules = s.SNCodeRules
	if model.SNCodeRules == 1 {
		model.SNMax = s.MaxSNCode
		model.SNMin = s.MinSNCode
	}

	//如果SN格式是带括号的，在SN上增加括号，以及在LOT号上也增加括号
	if model.SNFormat == 1 {
		model.SNMax = model.SNFormatInfo + model.SNMax
		model.SNMin = model.SNFormatInfo + model.SNMin
		model.BatchCode = model.SNFormatInfo + model.BatchCode
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
	if batch.SNCodeRules == 0 { //自动生成SN号，才需要批量添加SN信息
		var err error
		var number int = batch.BatchNumber + batch.BatchExtra
		for i := 1; i < number+1; i++ {
			var data models.SNInfo
			data.BatchId = batch.BatchId
			data.BatchName = batch.BatchName
			data.BatchCode = batch.BatchCode
			data.WorkCode = batch.WorkCode
			data.ProductCode = batch.ProductCode
			data.UDI = batch.UDI
			data.ProductMonth = batch.ProductMonth
			data.ProductId = batch.ProductId
			date := data.ProductMonth

			year := date.Year()
			ycode := (year - 33) % 100
			month := date.Month()
			mcode := month + 22
			sn := fmt.Sprintf("%06d", i)
			data.SNCode = strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + data.ProductCode + sn

			/*
				External := batch.External
				if External == 1 {
					data.SNCode = "(01)" + data.SNCode
				}*/

			//如果SN格式是带括号的，在SN上增加括号
			if batch.SNFormat == 1 {
				data.SNCode = batch.SNFormatInfo + data.SNCode
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

// UpdateSNInfoStatus 修改SN状态
func (e *BatchInfo) UpdateSNInfoStatus(c *dto.SNInfoUpdateReq) error {
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

func (e *BatchInfo) GenerateUpdateID(model *models.BatchInfo, s *dto.BatchInfoUpdateReq) error {
	var list []models.BatchInfo
	//	e.Orm.Unscoped().Where("product_id= ? AND DATE_FORMAT(product_month,'%Y-%m-&d ')= ?", s.ProductId, model.ProductMonth).Find(&list)
	e.Orm.Unscoped().Where("product_id= ? AND product_month= ?", model.ProductId, model.ProductMonth).Find(&list)

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
	date, _ := time.Parse("2006-01-02", s.ProductMonth+"-03")
	if !isLast {
		if model.BatchCodeFormat != s.BatchCodeFormat || model.SNCodeRules != s.SNCodeRules || model.ProductId != s.ProductId || model.BatchNumber+model.BatchExtra != s.BatchNumber+s.BatchExtra || date.Year() != model.ProductMonth.Year() || date.Month() != model.ProductMonth.Month() {
			e.Log.Info("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v", model.BatchCodeFormat, s.BatchCodeFormat, model.SNCodeRules, s.SNCodeRules, model.ProductId, uint(s.ProductId), model.BatchNumber+model.BatchExtra, s.BatchNumber+s.BatchExtra, date, model.ProductMonth)
			return errors.New("不是当月最后一批，只能修改SN格式，批次状态，工单号，图样，备注等信息")
		}
	}
	model.BatchName = s.BatchName
	model.BatchNumber = s.BatchNumber
	model.BatchExtra = s.BatchExtra
	e.Orm.Unscoped().Where("product_id= ? AND DATE_FORMAT(product_month,'%Y-%m')= ?", s.ProductId, s.ProductMonth).Find(&list)
	var autoSNSum int = 0
	var autoBatchCount int = 1
	for _, batch := range list {
		if batch.SNCodeRules == 0 && batch.BatchId < model.BatchId {
			autoSNSum = autoSNSum + batch.BatchNumber + batch.BatchExtra
		}
		if batch.BatchCodeFormat == 0 && batch.BatchId < model.BatchId {
			autoBatchCount++
		}
	}
	if isLast {
		year := date.Year()
		ycode := (year - 33) % 100
		month := date.Month()
		mcode := month + 22
		smin := fmt.Sprintf("%06d", autoSNSum+1)
		smax := fmt.Sprintf("%06d", autoSNSum+s.BatchNumber+s.BatchExtra)
		model.SNMax = strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + s.ProductCode + smax
		model.SNMin = strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + s.ProductCode + smin
		model.ProductMonth = date
		var cstr string = fmt.Sprintf("%03d", autoBatchCount)
		monthStr := fmt.Sprintf("%02d", int(month))
		model.BatchCode = strconv.Itoa(year) + monthStr + cstr
		model.SNFormat = s.SNFormat
		model.SNFormatInfo = s.SNFormatInfo

		model.ProductId = s.ProductId
		model.ProductCode = s.ProductCode
		//是否手动填写LOT号
		model.BatchCodeFormat = s.BatchCodeFormat

		if model.BatchCodeFormat == 1 {
			model.BatchCode = s.BatchCodeInfo
			model.BatchCodeFormatInfo = s.BatchCodeInfo
		}

		//客户指定SN号
		model.SNCodeRules = s.SNCodeRules
		if model.SNCodeRules == 1 {
			model.SNMax = s.MaxSNCode
			model.SNMin = s.MinSNCode
		}

		//如果SN格式是带括号的，在SN上增加括号，以及在LOT号上也增加括号
		if model.SNFormat == 1 {
			model.SNMax = model.SNFormatInfo + model.SNMax
			model.SNMin = model.SNFormatInfo + model.SNMin
			model.BatchCode = model.SNFormatInfo + model.BatchCode
		}
	} else {
		if model.SNFormat == 0 && s.SNFormat == 1 {
			model.SNFormat = s.SNFormat
			model.SNFormatInfo = s.SNFormatInfo
			model.SNMax = model.SNFormatInfo + model.SNMax
			model.SNMin = model.SNFormatInfo + model.SNMin
			model.BatchCode = model.SNFormatInfo + model.BatchCode
		} else if model.SNFormat == 1 && s.SNFormat == 0 {
			model.SNFormat = s.SNFormat
			model.SNFormatInfo = ""
			model.SNMax = string([]rune(model.SNMax)[len([]rune(model.SNFormatInfo)):len([]rune(model.SNMax))])
			model.SNMin = string([]rune(model.SNMin)[len([]rune(model.SNFormatInfo)):len([]rune(model.SNMin))])
			model.BatchCode = string([]rune(model.BatchCode)[len([]rune(model.SNFormatInfo)):len([]rune(model.BatchCode))])
		}
	}
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
