// Package db provides ...
package db

import "mysslee_qcloud/model"

var mProduct = model.Product{}

// 获取产品
func GetProductById(id int) (*model.Product, error) {
	p := new(model.Product)
	err := gormDB.Where("id=?", id).First(p).Error
	return p, err
}

// 获取产品
func GetProductByPid(pid int) (*model.Product, error) {
	p := new(model.Product)
	err := gormDB.Where("pid=?", pid).First(p).Error
	return p, err
}

// 获取产品列表
func GetProductList() ([]*model.Product, error) {
	var products []*model.Product
	err := gormDB.Find(&products).Error
	return products, err
}

// 添加产品
func AddProduct(product *model.Product) error {
	return gormDB.Create(product).Error
}

// 是否存在
func IsExistProduct(id int) bool {
	var count int
	err := gormDB.Model(mProduct).
		Where("id=?", id).
		Count(&count).Error
	return err == nil && count > 0
}

// 是否存在
func IsExistProductByPid(pid int) bool {
	var count int
	err := gormDB.Model(mProduct).
		Where("pid=?", pid).
		Count(&count).Error
	return err == nil && count > 0
}
