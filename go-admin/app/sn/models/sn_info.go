package models

import (
	"go-admin/common/models"
	"time"
)

type SNInfo struct {
	SNId     	int    `gorm:"primaryKey;autoIncrement" json:"SNId"` //SNId
	SNCode 		string `gorm:"size:128;" json:"SNCode"` //SN码
	BatchId     int    `gorm:"size:19;" json:"BatchId"` //批次ID
	BatchName   string `gorm:"size:128;" json:"BatchName"`              //批次名称
	BatchCode   string `gorm:"size:128;" json:"BatchCode"`
	WorkCode    string `gorm:"size:128;" json:"WorkCode"`    //岗位代码
	ProductCode string `gorm:"size:128;" json:"ProductCode"` //产品型号
	UDI         string `gorm:"size:128;" json:"UDI"`         //UDI
	Status  int    `gorm:"size:4;" json:"status"`    //状态

	models.ControlBy
	models.ModelTime

	ProductMonth time.Time `json:"ProductMonth" gorm:"column(product_month);comment:批次月份"`
}

func (SNInfo) TableName() string {
	return "sn_info"
}

func (e *SNInfo) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *SNInfo) GetId() interface{} {
	return e.SNId
}
