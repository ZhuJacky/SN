// Package db provides ...
package db

import (
	"mysslee_qcloud/model"

	"github.com/jinzhu/gorm"
)

var mLimitConsume = model.LimitConsume{}

// IncrLimitConsume 消耗额度
func IncrLimitConsume(uin, date string, limit int) (ok bool, err error) {
	tx := gormDB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	consume := new(model.LimitConsume)
	err = tx.Set("gorm:query_option", "FOR UPDATE").
		Where("uin=? AND date=?", uin, date).First(consume).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return false, err
		}
		consume = &model.LimitConsume{
			Uin:  uin,
			Date: date,
			Cost: 1,
		}
		err = tx.Create(consume).Error
		return true, err
	}
	// 消耗
	if consume.Cost >= limit {
		return false, nil
	}
	err = tx.Model(mLimitConsume).Where("uin=? AND date=?", uin, date).
		Update("cost", consume.Cost+1).Error
	if err != nil {
		return false, err
	}

	return true, nil
}
