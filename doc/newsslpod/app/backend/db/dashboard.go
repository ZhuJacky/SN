package db

import (
	"mysslee_qcloud/model"

	"github.com/jinzhu/gorm"
)

var modelDashboardResult = model.DashboardResult{}

//根据uid获取dashboard的结果
func GetDashboardResultByAccountId(accountId int) (data string, err error) {
	result := new(model.DashboardResult)
	err = gormDB.Model(&model.DashboardResult{}).Select("result").Where("account_id=?", accountId).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return "null", nil
	}
	return result.Result, err
}

//更新或插入用户的dashboard展示信息
func UpdateDashBoardResultByUid(accountId int, dashboard string) error {
	var count int
	err := gormDB.Model(modelDashboardResult).Where("account_id=?", accountId).Count(&count).Error
	if err != nil {
		return err
	}
	// create record
	if count == 0 {
		result := &model.DashboardResult{
			AccountId: accountId,
			Result:    dashboard,
		}
		return gormDB.Create(result).Error
	}

	return gormDB.Model(modelDashboardResult).
		Where("account_id=?", accountId).Update("result", dashboard).Error
}
