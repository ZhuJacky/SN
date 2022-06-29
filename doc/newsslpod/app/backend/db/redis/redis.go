// Package db provides ...
package redis

import (
	"fmt"
	"time"

	"mysslee_qcloud/redis"
	"mysslee_qcloud/utils"
)

var (
	EmailExpire     = time.Hour * 24
	TXTRecordExpire = time.Hour * 1
	PhoneExpire     = time.Minute * 5
	OtpExpire       = time.Minute * 5
)

// 记录发送冷却期
func SetWarnNotice(key string, value int64) error {
	return redis.RedisCli.Set("warnnotice:"+key, value, time.Hour*3).Err()
}

// 获取告警冷却
func IsExistWarnNotice(key string) bool {
	count := redis.RedisCli.Exists("warnnotice:" + key).Val()
	return count > 0
}

// 唯一码生成
var (
	inviteCode int64 = 0xac12
)

func InitUniqueId() error {
	if redis.RedisCli.Exists("uniqueId").Val() == 0 {
		return redis.RedisCli.Set("uniqueId", inviteCode, 0).Err()
	}
	return nil
}

func UniqueOrderId() (string, error) {
	id, err := redis.RedisCli.Incr("uniqueId").Result()
	code := utils.RandomCode(4, true)
	return fmt.Sprintf("%s%s%d", time.Now().Format("20060102150405"), code, id), err
}

func UniqueResrouceId() (string, error) {
	id, err := redis.RedisCli.Incr("uniqueId").Result()
	code := utils.RandomCode(6, false)
	return fmt.Sprintf("plan-%s%d", code, id), err
}

// for test
func DelUniqueId() error {
	return redis.RedisCli.Del("uniqueId").Err()
}

// 每月额度限制
func SetNoticeLimit(aid int, name string, sent int) (err error) {
	now := time.Now()
	exp := utils.MonthToEndDuration(now)
	return redis.RedisCli.Set(fmt.Sprintf("%s:%d", name, aid), sent, exp).Err()
}

func GetNoticeLimit(uin, name string) (int, error) {
	// if name != model.EmailLimitName ||
	// 	name != model.WechatLimitName ||
	// 	name != model.PhoneLimitName ||
	// 	name != model.MaxAddLimitName ||
	// 	name != model.MaxMonitorLimitName {
	// 	return 0, errors.New("redis: unknown set limit name:", name)
	// }
	sent, err := redis.RedisCli.Get(fmt.Sprintf("%s:%s", name, uin)).Int64()

	if err != nil && redis.Nil == err {
		return 0, nil
	}
	return int(sent), err
}

// 缓存套餐额度信息
func SetPlanLimit(uin string, data []byte, d time.Duration) error {
	return redis.RedisCli.Set(fmt.Sprintf("planLimit:%s", uin), data, d).Err()
}

// 获取额度信息
func GetPlanLimit(uin string) ([]byte, error) {
	return redis.RedisCli.Get(fmt.Sprintf("planLimit:%s", uin)).Bytes()
}

// 清除额度信息
func DelPlanLimit(uin string) error {
	return redis.RedisCli.Del(fmt.Sprintf("planLimit:%s", uin)).Err()
}

// SetCertInfo 将证书信息写缓存
func SetCertInfo(hash string) {

}
