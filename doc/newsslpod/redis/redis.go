// Package redis provides ...
package redis

import (
	"time"

	"mysslee_qcloud/config"
	"mysslee_qcloud/l5"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

var (
	RedisCli *redis.Client
	Nil      = redis.Nil
)

func Init(app AppName) {
	if config.Conf.ForceL5 {
		srv, err := l5.GetRedisServer()
		if err != nil {
			panic(err)
		}
		RedisCli = redis.NewClient(&redis.Options{
			Addr:     srv.String(),
			Password: "",
			DB:       0,
		})
	} else {
		RedisCli = redis.NewClient(&redis.Options{
			Addr:     config.Conf.Redis.Address,
			Password: config.Conf.Redis.Password,
			DB:       0,
		})
	}
	_, err := RedisCli.Ping().Result()
	if err != nil {
		panic(err)
	}

	// 不用redis做服务发现了
	// checker fastcheck 还是得用redis
	if app == CheckerApp {
		go keepOnline(app)
	}
}

func keepOnline(app AppName) {
	t := time.NewTicker(time.Second * 2)
	for range t.C {
		selfIP, err := selfIPAddress()
		if err != nil {
			logrus.Error("keepOnline.selfIPAddress: ", err)
			continue
		}

		err = RedisCli.Set(string(app)+selfIP, selfIP, time.Second*3).Err()
		if err != nil {
			logrus.Error("keepOnline.Set: ", err)
			continue
		}
		err = RedisCli.SAdd(string(app), selfIP).Err()
		if err != nil {
			logrus.Error("keepOnline.SAdd: ", err)
			continue
		}
	}
}

// ScanApp return node ip
func ScanApp(app AppName) ([]string, error) {
	mems, err := RedisCli.SMembers(string(app)).Result()
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, ip := range mems {
		ip, err := RedisCli.Get(string(app) + ip).Result()
		if err != nil {
			if err == redis.Nil {
				RedisCli.SRem(string(app), ip).Err()
				continue
			}
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, nil
}

// Lock lock key
func Lock(key string) bool {
	ok, err := RedisCli.SetNX(key, "", time.Second*5).Result()
	if err != nil || !ok {
		return false
	}
	return true
}

// Unlock unlock key
func Unlock(key string) error {
	return RedisCli.Del(key).Err()
}
