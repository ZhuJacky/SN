// Package db provides ...
package db

import "mysslee_qcloud/model"

var modelAccount = model.Account{}

// GetAccountById get account by id
func GetAccountById(id int) (*model.Account, error) {
	u := new(model.Account)
	err := gormDB.Where("id=?", id).First(u).Error
	return u, err
}

// UpdateAccount update account specified fields
func UpdateAccount(aid int, fields map[string]interface{}) error {
	return gormDB.Model(modelAccount).
		Where("id=?", aid).Updates(fields, false).Error
}

// SetAccountAggrFlag set account aggregate flag
func SetAccountAggrFlag(domainId int) error {
	return gormDB.Exec("UPDATE accounts JOIN account_domains SET aggregate=aggregate+1"+
		" WHERE account_domains.account_id=accounts.id AND account_domains.domain_id=?", domainId).Error
}
