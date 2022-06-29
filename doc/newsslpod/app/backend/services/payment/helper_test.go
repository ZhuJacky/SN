// Package payment provides ...
package payment

import (
	"testing"
	"time"
)

func TestParseAddTime(t *testing.T) {
	t2 := parseAddTime(time.Now(), "y", 1)
	t.Log(t2)
}
