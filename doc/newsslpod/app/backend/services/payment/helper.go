// Package payment provides ...
package payment

import (
	"time"

	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/model"
)

func checkProductId(productId int) bool {
	return productId/1000 == 2
}

func checkPid(pid int) bool {
	return db.IsExistProductByPid(pid)
}

func checkTimeSpan(timeSpan int) bool {
	return timeSpan <= 3 && timeSpan >= 1
}

func checkTimeUnit(timeUnit string, need string) bool {
	return timeUnit == need
}

func checkGoodsNum(num int) bool {
	return num == 1
}

func checkDeadline(deadline string, need time.Time) bool {
	return deadline == need.Format(model.TIME_FORMAT)
}

func checkExtendTime(extend int) bool {
	return extend > 365*24*3600 && extend <= 10*365*24*3600
}

func checkAppId(appid int) bool {
	return appid > 0
}

func parseAddTime(t time.Time, timeUnit string, timeSpan int) time.Time {
	switch timeUnit {
	case "y":
		return t.AddDate(timeSpan, 0, 0)
	case "m":
		return t.AddDate(0, timeSpan, 0)
	case "d":
		return t.AddDate(0, 0, timeSpan)
	case "h":
		return t.Add(time.Hour * time.Duration(timeSpan))
	case "M":
		return t.Add(time.Minute * time.Duration(timeSpan))
	case "s":
		return t.Add(time.Second * time.Duration(timeSpan))
	}
	return t
}

func checkSSLPod(sslpod int) bool {
	return sslpod == 1
}
