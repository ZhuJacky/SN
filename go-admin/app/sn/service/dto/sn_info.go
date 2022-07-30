package dto

import (
	"go-admin/app/sn/models"
	"go-admin/common/dto"
	common "go-admin/common/models"
)

// SysPostPageReq 列表或者搜索使用结构体
type SNInfoPageReq struct {
	dto.Pagination `search:"-"`
	BatchId      string `form:"postCode" search:"type:exact;column:batch_id;table:sn_info" comment:"id"` // id
	Status       int    `form:"status" search:"type:exact;column:status;table:sn_info" comment:"status"`   // status
}

func (m *SNInfoPageReq) GetNeedSearch() interface{} {
	return *m
}

// SysPostUpdateReq 改使用的结构体
type SNInfoUpdateReq struct {
	SNId   int `uri:"id"  comment:"id"`
	Status int `form:"status"   comment:"状态"`
	common.ControlBy
}

func (s *SNInfoUpdateReq) GetId() interface{} {
	return s.SNId
}

// SysPostGetReq 获取单个的结构体
type SNInfoGetReq struct {
	Id int `uri:"id"`
}

func (s *SNInfoGetReq) GetId() interface{} {
	return s.Id
}

// SysPostDeleteReq 删除的结构体
type SNInfoDeleteReq struct {
	Ids []int `json:"ids"`
	common.ControlBy
}

func (s *SNInfoDeleteReq) Generate(model *models.BatchInfo) {
	if s.ControlBy.UpdateBy != 0 {
		model.UpdateBy = s.UpdateBy
	}
	if s.ControlBy.CreateBy != 0 {
		model.CreateBy = s.CreateBy
	}
}

func (s *SNInfoDeleteReq) GetId() interface{} {
	return s.Ids
}
