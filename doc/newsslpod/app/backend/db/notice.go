// Package db provides ...
package db

import "mysslee_qcloud/model"

// 创建
func AddOrUpNoticeInfo(nInfo *model.NoticeInfo) error {
	if nInfo.Id > 0 {
		return gormDB.Save(nInfo).Error
	}
	return gormDB.Create(nInfo).Error
}

// 获取通知开关信息
func GetNoticeInfo(uin string) (*model.NoticeInfo, error) {
	nInfo := new(model.NoticeInfo)
	err := gormDB.Where("uin = ?", uin).First(nInfo).Error
	return nInfo, err
}
