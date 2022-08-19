package dto

import (
	"go-admin/common/dto"
)

type BoxRelationInfoPageReq struct {
	dto.Pagination `search:"-"`
	BoxId          int `form:"BoxId" search:"type:exact;column:box_id;table:sn_box_relation" comment:"box_id"` //BoxId
}

func (m *BoxRelationInfoPageReq) GetNeedSearch() interface{} {
	return *m
}

type BoxRelationInfoResultObj struct {
	Status int `comment:"状态"`
	BoxId  int `comment:"箱号"`
}
