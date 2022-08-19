package models

import (
	"go-admin/common/models"
)

type SNBoxRelation struct {
	BoxRelationId int    `gorm:"primaryKey;autoIncrement" json:"BoxRelationId"`
	BoxId         int    `gorm:"size:19;" json:"BoxId"`
	SNCode        string `gorm:"size:128;" json:"SNCode"`
	ScanSource    string `gorm:"size:128;" json:"ScanSource"`
	BatchCode     string `gorm:"size:128;" json:"BatchCode"`
	ProductCode   string `gorm:"size:128;" json:"ProductCode"` //产品型号
	BoxSum        int    `gorm:"size:4;" json:"BoxSum"`        //装箱数量
	models.ControlBy
	models.ModelTime
}

func (SNBoxRelation) TableName() string {
	return "sn_box_relation"
}

func (e *SNBoxRelation) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *SNBoxRelation) GetId() interface{} {
	return e.BoxRelationId
}
