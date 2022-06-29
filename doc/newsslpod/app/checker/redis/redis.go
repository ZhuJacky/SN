// Package redis provides ...
package redis

import (
	"fmt"
	"time"

	"mysslee_qcloud/redis"
	"mysslee_qcloud/utils"
)

// 是否应该通知
func ShouldWarnNotice(key string) bool {
	return redis.RedisCli.Exists("warnnotice:"+key).Val() > 0
}

// 设置通知时间
func SetWarnNotice(key string, value int64) error {
	return redis.RedisCli.Set("warnnotice:"+key, value, time.Hour*3).Err()
}

func GetNoticeLimit(uin string, name string, allow int) (bool, error) {
	key := fmt.Sprintf("%s:%s", name, uin)
	sent, err := redis.RedisCli.Get(key).Int64()
	if err != nil && err != redis.Nil {
		return false, err
	}
	if int(sent) >= allow {
		return false, nil
	}
	now := time.Now()
	exp := utils.MonthToEndDuration(now)

	pipe := redis.RedisCli.Pipeline()
	pipe.Incr(key)
	pipe.Expire(key, exp)
	_, err = pipe.Exec()
	return true, err
}

// 获取额度信息
func GetPlanLimit(uin string) ([]byte, error) {
	return redis.RedisCli.Get("planLimit:" + uin).Bytes()
}
