// Package db provides ...
package db

import (
	"mysslee_qcloud/model"
)

var mAccount = &model.Account{}

// 添加
func AddAccount(a *model.Account) error {
	return gormDB.Create(a).Error
}

// NOTE 删除，供测试使用
func DelAccount(s string) error {
	return gormDB.Where("email=? OR phone=?", s).Delete(mAccount).Error
}

// 更新
func UpAccount(a *model.Account) error {
	return gormDB.Model(a).Updates(a).Error
}

func UpAccountFiled(a *model.Account, m map[string]interface{}) error {
	return gormDB.Model(a).Updates(m).Error
}

// 查询
func GetAccountById(id int) (*model.Account, error) {
	u := new(model.Account)
	err := gormDB.Where("id=?", id).First(u).Error
	return u, err
}

// 获取账户仅注册的
func GetAccountByUin(uin string) (*model.Account, error) {
	u := new(model.Account)
	err := gormDB.First(u, "uin=?", uin).Error
	return u, err
}

// 获取用户非拉黑
func GetAccountNormal(s string) (*model.Account, error) {
	u := new(model.Account)
	err := gormDB.First(u, "email=? OR phone=? AND status!=?", s, s, model.StatusForbidden).Error
	return u, err
}

// 获取账号数量
func GetAccountCount() (int, error) {
	var count int
	err := gormDB.Model(mAccount).Count(&count).Error
	return count, err
}

// SetAccountAggrFlagOnly set account aggregate flag only specified account
func SetAccountAggrFlagOnly(aid int) error {
	return gormDB.Exec("UPDATE accounts SET aggregate=aggregate+1 where id=?", aid).Error
}
