// Package redis provides ...
package redis

import (
	"testing"

	"mysslee_qcloud/redis"
)

func init() {
	redis.Init(redis.CheckerApp)
}

func TestGetNoticeLimit(t *testing.T) {
	ok, err := GetNoticeLimit("1", "limit_test", 10)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ok)
}
