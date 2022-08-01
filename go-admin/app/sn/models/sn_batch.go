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

	Status        int    `gorm:"size:4;" json:"status"`         //状态
	Comment       string `gorm:"size:255;" json:"Comment"`      //描述备注
	SNFormat      int    `gorm:"size:4;" json:"SNFormat"`       //SN格式
	SNFormatInfo  string `gorm:"size:255;" json:"SNFormatInfo"` //SN格式信息
	UDIFormatInfo string `gorm:"size:255;" json:"UDIFormatInfo"`
	LOTFormatInfo string `gorm:"size:255;" json:"LOTFormatInfo"`

	BatchCodeFormat     int    `gorm:"size:4;" json:"BatchCodeFormat"`       //批号格式
	BatchCodeFormatInfo string `gorm:"size:255;" json:"BatchCodeFormatInfo"` //SN格式信息
	SNCodeRules         int    `gorm:"size:4;" json:"SNCodeRules"`           //SN生成规则
	BatchImgFile        string `gorm:"size:255;" json:"BatchImgFile"`        //
	External            int    `gorm:"size:4;" json:"External"`              //制作类型
	AutoSNSum           int    `gorm:"size:4;" json:"autoSNSum"`             //SN启始号
	models.ControlBy
	models.ModelTime

	ProductMonth time.Time `json:"ProductMonth" gorm:"column(product_month);comment:批次月份"`

	ProductId uint         `gorm:"column(product_id)" json:"ProductId"` //批次ID
	Product   *ProductInfo `json:"Product"`
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
