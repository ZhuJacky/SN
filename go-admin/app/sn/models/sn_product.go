package models

import (
	"go-admin/common/models"
)

type ProductInfo struct {
	ProductId   int    `gorm:"primaryKey;autoIncrement" json:"ProductId"` //批次ID
	ProductName string `gorm:"size:128;" json:"ProductName"`              //批次名称
	ProductCode string `gorm:"size:128;" json:"ProductCode"`              //产品型号
	UDI         string `gorm:"size:128;" json:"UDI"`                      //

	Status  int    `gorm:"size:4;" json:"status"`    //状态
	Comment string `gorm:"size:255;" json:"Comment"` //描述备注

	models.ControlBy
	models.ModelTime

	ImageFile string `gorm:"size:128;column(image_file);" json:"ImageFile"`
}

func (ProductInfo) TableName() string {
	return "sn_product_info"
}

func (e *ProductInfo) Generate() models.ActiveRecord {
	o := *e
	return &o
}

func (e *ProductInfo) GetId() interface{} {
	return e.ProductId
}
