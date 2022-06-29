// Package db provides ...
package db

import (
	"testing"
	"time"
)

func TestGetNeedCheckDomains(t *testing.T) {
	list, err := GetNeedCheckDomains(time.Now(), func(total int) (int, error) {
		return 10, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("list==", len(list))
}
