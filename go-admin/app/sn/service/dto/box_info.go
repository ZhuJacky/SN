package dto

import (
	"go-admin/common/dto"
	common "go-admin/common/models"
)

type BoxInfoPageReq struct {
	dto.Pagination `search:"-"`
	BatchCode      string `form:"batchCode" search:"type:exact;column:batch_code;table:sn_box_info" comment:"batch_code"`       // batch_code
	ProductCode    string `form:"productCode" search:"type:exact;column:product_code;table:sn_box_info" comment:"product_code"` // product_code
}

type BoxInfoUpdateReq struct {
	BoxId  int `uri:"BoxId"  comment:"box_id"`
	BoxSum int `form:"BoxSum"   comment:"box_sum"`
	common.ControlBy
}

func (s *BoxInfoUpdateReq) GetId() interface{} {
	return s.BoxId
}

func (m *BoxInfoPageReq) GetNeedSearch() interface{} {
	return *m
}
