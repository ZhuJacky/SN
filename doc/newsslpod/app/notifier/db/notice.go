// Package db provides ...
package db

import (
	"errors"
	"time"

	"mysslee_qcloud/config"
	"mysslee_qcloud/model"
	"mysslee_qcloud/redis"
)

const key = "sslpod:notify"

var modelNoticeMsg = model.NoticeMsg{}

// 获取通知
func GetNoticeMsgs(page, pageSize int) (msgs []*model.NoticeMsg, err error) {
	// redis 并发锁
	count := 0
	for {
		count++
		if count > 5 {
			return nil, errors.New("unlocked key: need-notice-lock")
		}
		if redis.Lock("need-notice-lock") {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
	defer redis.Unlock("need-notice-lock")

	// notice limit
	val, err := redis.RedisCli.Get(key).Int()
	if err != nil {
		if err != redis.Nil {
			return nil, err
		}
		err = redis.RedisCli.Set(key, 0, time.Hour).Err()
		if err != nil {
			return nil, err
		}
	}
	if val >= config.Conf.Notifier.RateLimit {
		return nil, errors.New("too fast")
	}
	ttl, err := redis.RedisCli.TTL(key).Result()
	if err != nil {
		return nil, err
	}
	db := gormDB.Model(modelNoticeMsg).Where("noticed_at IS NULL")
	// get count
	err = db.Count(&count).Error
	if err != nil || count == 0 {
		return nil, err
	}
	if count < pageSize {
		pageSize = count
	}
	if val+pageSize > config.Conf.Notifier.RateLimit {
		pageSize = config.Conf.Notifier.RateLimit - val
	}
	// update etcd
	err = redis.RedisCli.Set(key, val+pageSize, ttl).Err()
	if err != nil {
		return nil, err
	}
	// select data
	var temp []*model.NoticeMsg
	err = db.Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&temp).Error
	if err != nil || len(temp) == 0 {
		return nil, err
	}
	var ids []int
	for _, v := range temp {
		msgs = append(msgs, v)
		ids = append(ids, v.Id)
	}
	// update database
	err = gormDB.Model(modelNoticeMsg).
		Where("id IN (?)", ids).
		Update("noticed_at", time.Now()).Error
	return msgs, err
}

// 更新通知
func UpNoticeMsg(id int, fields map[string]interface{}) error {
	return gormDB.Model(modelNoticeMsg).
		Where("id=?", id).Updates(fields).Error
}

// 创建
func AddOrUpNoticeInfo(nInfo *model.NoticeInfo) error {
	if nInfo.Id > 0 {
		return gormDB.Save(nInfo).Error
	}
	return gormDB.Create(nInfo).Error
}

// 添加通知
func AddNoticeMsg(msg *model.NoticeMsg) error {
	return gormDB.Create(msg).Error
}
