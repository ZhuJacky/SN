// Package db provides ...
package db

import (
	"encoding/json"
	"testing"
)

func TestGetProductByPid(t *testing.T) {
	p, err := GetProductByPid(15958)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(p)
}

func TestGetProductList(t *testing.T) {
	list, err := GetProductList()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}
