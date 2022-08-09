package models

import (
	"go-admin/common/models"
)

type SNBoxInfo struct {
	BoxId       int    `gorm:"primaryKey;autoIncrement" json:"BoxId"`
	BatchId     int    `gorm:"size:19;" json:"BatchId"` //批次ID
	BatchCode   string `gorm:"size:128;" json:"BatchCode"`
	WorkCode    string `gorm:"size:128;" json:"WorkCode"`    //岗位代码
	ProductCode string `gorm:"size:128;" json:"ProductCode"` //产品型号
	ScanSource  string `gorm:"size:128;" json:"ScanSource"`  //产品型号
	UDI         string `gorm:"size:128;" json:"UDI"`         //UDI
	Status      int    `gorm:"size:4;" json:"Status"`        //状态
	BoxSum      int    `gorm:"size:4;" json:"BoxSum"`        //状态
	models.ControlBy
	models.ModelTime
}

func (SNBoxInfo) TableName() string {
	return "sn_box_info"
}

func (e *SNBoxInfo) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *SNBoxInfo) GetId() interface{} {
	return e.BoxId
}
