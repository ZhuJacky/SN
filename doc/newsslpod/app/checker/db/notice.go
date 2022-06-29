// Package db provides ...
package db

import "mysslee_qcloud/model"

var modelNoticeInfo = model.NoticeInfo{}

// 获取通知设置
func GetNoticeInfo(uin string) (*model.NoticeInfo, error) {
	nInfo := new(model.NoticeInfo)
	err := gormDB.Where("uin = ?", uin).First(nInfo).Error
	return nInfo, err
}

// 添加通知
func AddNoticeMsg(msg *model.NoticeMsg) error {
	return gormDB.Create(msg).Error
}
