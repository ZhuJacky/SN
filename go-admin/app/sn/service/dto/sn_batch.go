package dto

import (
	"go-admin/app/sn/models"
	common "go-admin/common/models"
	"time"

	"go-admin/common/dto"
)

// SysPostPageReq 列表或者搜索使用结构体
type BatchInfoPageReq struct {
	dto.Pagination `search:"-"`
	BatchId        int    `form:"postId" search:"type:exact;column:batch_id;table:batch_info" comment:"id"`        // id
	BatchName      string `form:"postName" search:"type:contains;column:batch_name;table:batch_info" comment:"名称"` // 名称
	BatchCode      string `form:"postCode" search:"type:contains;column:batch_code;table:batch_info" comment:"编码"` // 编码

	ProductCode string `form:"productCode" search:"type:contains;column:product_code;table:batch_info" comment:"编码"` // 编码
	SNMax       string `form:"snMax" search:"type:exact;column:snmax;table:batch_info" comment:"SNMAX"`              // 编码
	SNMin       string `form:"snMin" search:"type:exact;column:snmax;table:batch_info" comment:"SNMIN"`              // 编码
	Status      int    `form:"status" search:"type:exact;column:status;table:batch_info" comment:"状态"`               // 状态
	Comment     string `form:"remark" search:"type:exact;column:comment;table:batch_info" comment:"备注"`              // 备注
}

func (m *BatchInfoPageReq) GetNeedSearch() interface{} {
	return *m
}

// SysPostInsertReq 增使用的结构体
type BatchInfoInsertReq struct {
	PostId       int       `uri:"id"  comment:"id"`
	PostName     string    `form:"postName"  comment:"名称"`
	PostCode     string    `form:"postCode" comment:"编码"`
	Status       int       `form:"status"   comment:"状态"`
	Remark       string    `form:"remark"   comment:"备注"`
	ProductMonth time.Time `form:"productDate"   comment:"备注"`
	common.ControlBy
}

func (s *BatchInfoInsertReq) Generate(model *models.BatchInfo) {
	model.BatchName = s.PostName
	model.BatchCode = s.PostCode
	model.Status = s.Status
	model.Comment = s.Remark
	model.ProductMonth = time.Now()
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
	if s.ControlBy.CreateBy != 0 {
		model.CreateBy = s.CreateBy
	}
}

// GetId 获取数据对应的ID
func (s *BatchInfoInsertReq) GetId() interface{} {
	return s.PostId
}

// SysPostUpdateReq 改使用的结构体
type BatchInfoUpdateReq struct {
	PostId   int    `uri:"id"  comment:"id"`
	PostName string `form:"postName"  comment:"名称"`
	PostCode string `form:"postCode" comment:"编码"`
	Sort     int    `form:"sort" comment:"排序"`
	Status   int    `form:"status"   comment:"状态"`
	Remark   string `form:"remark"   comment:"备注"`
	common.ControlBy
}

func (s *BatchInfoUpdateReq) Generate(model *models.BatchInfo) {
	model.BatchId = s.PostId
	model.BatchName = s.PostName
	model.BatchCode = s.PostCode
	model.Status = s.Status
	model.Comment = s.Remark
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
	if s.ControlBy.CreateBy != 0 {
		model.CreateBy = s.CreateBy
	}
}

func (s *BatchInfoUpdateReq) GetId() interface{} {
	return s.PostId
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
