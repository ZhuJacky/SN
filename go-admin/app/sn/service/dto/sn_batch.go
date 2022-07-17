package dto

import (
	"fmt"
	"go-admin/app/sn/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
	"strconv"
	"time"
)

// SysPostPageReq 列表或者搜索使用结构体
type BatchInfoPageReq struct {
	dto.Pagination `search:"-"`
	BatchId        int    `form:"postId" search:"type:exact;column:batch_id;table:sn_batch_info" comment:"id"`        // id
	BatchName      string `form:"postName" search:"type:contains;column:batch_name;table:sn_batch_info" comment:"名称"` // 名称
	BatchCode      string `form:"postCode" search:"type:contains;column:batch_code;table:sn_batch_info" comment:"编码"` // 编码

	ProductCode string `form:"productCode" search:"type:contains;column:product_code;table:sn_batch_info" comment:"编码"` // 编码
	SNMax       string `form:"snMax" search:"type:exact;column:snmax;table:sn_batch_info" comment:"SNMAX"`              // 编码
	SNMin       string `form:"snMin" search:"type:exact;column:snmax;table:sn_batch_info" comment:"SNMIN"`              // 编码
	Status      int    `form:"status" search:"type:exact;column:status;table:sn_batch_info" comment:"状态"`               // 状态
	Comment     string `form:"Comment" search:"type:exact;column:comment;table:sn_batch_info" comment:"备注"`             // 备注
}

func (m *BatchInfoPageReq) GetNeedSearch() interface{} {
	return *m
}

// SysPostInsertReq 增使用的结构体
type BatchInfoInsertReq struct {
	BatchId     int    `uri:"id"  comment:"id"`
	BatchName   string `form:"BatchName"  comment:"批次名称"`
	BatchNumber int    `form:"BatchNumber"  comment:"批次数量"`
	BatchExtra  int    `form:"BatchExtra"  comment:"附加数量"`
	ProductCode string `form:"ProductCode" comment:"产品型号"`
	WorkCode    string `form:"WorkCode" comment:"工单号"`
	UDI         string `form:"UDI" comment:"UDI号"`

	Status       int    `form:"status"   comment:"状态"`
	Comment      string `form:"Comment"   comment:"备注"`
	ProductMonth string `form:"ProductMonth"   comment:"生产月份"`
	common.ControlBy
}

func (s *BatchInfoInsertReq) Generate(model *models.BatchInfo) {
	model.BatchName = s.BatchName
	model.BatchNumber = s.BatchNumber
	model.BatchExtra = s.BatchExtra
	model.ProductCode = s.ProductCode
	model.UDI = s.UDI
	model.WorkCode = s.WorkCode
	model.Status = s.Status
	model.Comment = s.Comment
	date, _ := time.Parse("2006-01-02", s.ProductMonth+"-03")
	model.ProductMonth = date
	year := date.Year()
	ycode := (year - 33) % 100
	month := date.Month()
	mcode := month + 22
	smin := fmt.Sprintf("%06d", 1)
	smax := fmt.Sprintf("%06d", model.BatchNumber+model.BatchExtra)
	model.SNMax = strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + smin
	model.SNMin = strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + smax
	model.BatchCode = strconv.Itoa(year) + strconv.Itoa(int(month)) + "001"

	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
	if s.ControlBy.CreateBy != 0 {
		model.CreateBy = s.CreateBy
	}
}

// GetId 获取数据对应的ID
func (s *BatchInfoInsertReq) GetId() interface{} {
	return s.BatchId
}

// SysPostUpdateReq 改使用的结构体
type BatchInfoUpdateReq struct {
	BatchId     int    `uri:"id"  comment:"id"`
	BatchName   string `form:"BatchName"  comment:"批次名称"`
	BatchNumber int    `form:"BatchNumber"  comment:"批次数量"`
	BatchExtra  int    `form:"BatchExtra"  comment:"附加数量"`
	ProductCode string `form:"ProductCode" comment:"产品型号"`
	WorkCode    string `form:"WorkCode" comment:"工单号"`
	UDI         string `form:"UDI" comment:"UDI号"`

	Status       int    `form:"status"   comment:"状态"`
	Comment      string `form:"Comment"   comment:"备注"`
	ProductMonth string `form:"ProductMonth"   comment:"生产月份"`
	common.ControlBy
}

func (s *BatchInfoUpdateReq) Generate(model *models.BatchInfo) {
	model.BatchName = s.BatchName
	model.BatchNumber = s.BatchNumber
	model.BatchExtra = s.BatchExtra
	model.ProductCode = s.ProductCode
	model.UDI = s.UDI
	model.WorkCode = s.WorkCode
	model.Status = s.Status
	model.Comment = s.Comment
	dateString := s.ProductMonth + "-03"
	//date, _ := time.Parse("2022-01-03", s.ProductMonth+"-03")
	date, _ := time.Parse("2006-01-02", dateString)
	model.ProductMonth = date

	year := date.Year()
	ycode := (year - 33) % 100
	month := date.Month()
	mcode := month + 22
	smin := fmt.Sprintf("%06d", 1)
	smax := fmt.Sprintf("%06d", model.BatchNumber+model.BatchExtra)
	model.SNMax = strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + smin
	model.SNMin = strconv.Itoa(ycode) + strconv.Itoa(int(mcode)) + smax
	model.BatchCode = strconv.Itoa(year) + strconv.Itoa(int(month)) + "001"

	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
	if s.ControlBy.CreateBy != 0 {
		model.CreateBy = s.CreateBy
	}
}

func (s *BatchInfoUpdateReq) GetId() interface{} {
	return s.BatchId
}

// SysPostGetReq 获取单个的结构体
type BatchInfoGetReq struct {
	Id int `uri:"id"`
}

func (s *BatchInfoGetReq) GetId() interface{} {
	return s.Id
}

// SysPostDeleteReq 删除的结构体
type BatchInfoDeleteReq struct {
	Ids []int `json:"ids"`
	common.ControlBy
}

func (s *BatchInfoDeleteReq) Generate(model *models.BatchInfo) {
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
	if s.ControlBy.CreateBy != 0 {
		model.CreateBy = s.CreateBy
	}
}

func (s *BatchInfoDeleteReq) GetId() interface{} {
	return s.Ids
}
