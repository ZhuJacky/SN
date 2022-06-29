// Package myconn provides ...
package myconn

import (
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	params := []int{2, 10, 20, 30}
	for i, v := range params {
		r := strings.NewReader("12345678901234567890")
		result, err := Read(r, v)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(i, string(result))
	}
}
