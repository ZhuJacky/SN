package dto

import (
	"go-admin/common/dto"
	common "go-admin/common/models"
)

type BoxInfoPageReq struct {
	dto.Pagination `search:"-"`
	BatchCode      string `form:"batchCode" search:"type:exact;column:batch_code;table:sn_box_info" comment:"batch_code"`       // batch_code
	ProductCode    string `form:"productCode" search:"type:exact;column:product_code;table:sn_box_info" comment:"product_code"` // product_code
	Status         int    `form:"status" search:"type:exact;column:status;table:sn_box_info" comment:"status"`                  // status
}

type BoxInfoUpdateStatusReq struct {
	BoxId  int `uri:"BoxId"  comment:"box_id"`
	Status int `form:"Status"   comment:"Status"`
	common.ControlBy
}

type BoxInfoUpdateReq struct {
	BoxId  int `uri:"BoxId"  comment:"box_id"`
	BoxSum int `form:"BoxSum"   comment:"box_sum"`
	common.ControlBy
}

func (s *BoxInfoUpdateReq) GetId() interface{} {
	return s.BoxId
}

func (s *BoxInfoUpdateStatusReq) GetId() interface{} {
	return s.BoxId
}

func (m *BoxInfoPageReq) GetNeedSearch() interface{} {
	return *m
}

type BoxInfoResultObj struct {
	Status int `comment:"状态"`
	BoxId  int `comment:"箱号"`
}
