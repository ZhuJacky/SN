package models

import (
	"go-admin/common/models"
	"time"
)

type BatchInfo struct {
	BatchId     int    `gorm:"primaryKey;autoIncrement" json:"BatchId"` //批次ID
	BatchName   string `gorm:"size:128;" json:"BatchName"`              //批次名称
	BatchCode   string `gorm:"size:128;" json:"BatchCode"`
	BatchNumber int    `gorm:"size:4;" json:"BatchNumber"`   //批次数量
	BatchExtra  int    `gorm:"size:4;" json:"BatchExtra"`    //批次备用数量
	WorkCode    string `gorm:"size:128;" json:"WorkCode"`    //岗位代码
	ProductCode string `gorm:"size:128;" json:"ProductCode"` //产品型号
	UDI         string `gorm:"size:128;" json:"UDI"`         //

	SNMax string `gorm:"column(SNMax);size:128;" json:"SNMax"`
	SNMin string `gorm:"column(SNMin);size:128;" json:"SNMin"`

	Status  int    `gorm:"size:4;" json:"status"`    //状态
	Comment string `gorm:"size:255;" json:"Comment"` //描述备注

	models.ControlBy
	models.ModelTime

	ProductMonth time.Time `json:"ProductMonth" gorm:"column(product_month);comment:批次月份"`
}

func (BatchInfo) TableName() string {
	return "sn_batch_info"
}

func (e *BatchInfo) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *BatchInfo) GetId() interface{} {
	return e.BatchId
}
