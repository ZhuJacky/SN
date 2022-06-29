// Package db provides ...
package db

import (
	"testing"
	"time"
)

func TestUpdateAccount(t *testing.T) {
	err := UpdateAccount(1, map[string]interface{}{
		"created_at": time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetAccountAggrFlag(t *testing.T) {
	err := SetAccountAggrFlag(1)
	if err != nil {
		t.Fatal(err)
	}
}
