package db

import "mysslee_qcloud/model"

// 查询
func GetAccountById(id int) (*model.Account, error) {
	u := new(model.Account)
	err := gormDB.Where("id=?", id).First(u).Error
	return u, err
}
