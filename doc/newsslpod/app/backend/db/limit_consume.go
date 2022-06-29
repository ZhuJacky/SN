// Package db provides ...
package db

import (
	"mysslee_qcloud/model"

	"github.com/jinzhu/gorm"
)

var mLimitConsume = &model.LimitConsume{}

// GetLimitConsume 获取额度消耗信息
func GetLimitConsume(uin, date string) (*model.LimitConsume, error) {
	consume := new(model.LimitConsume)

	err := gormDB.Where("uin=? AND date=?", uin, date).
		First(consume).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		consume = &model.LimitConsume{
			Uin:  uin,
			Date: date,
			Cost: 0,
		}
		err = gormDB.Create(consume).Error
	}
	return consume, err
}
