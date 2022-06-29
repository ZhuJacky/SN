// Package qcloud provides ...
package qcloud

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var req = &Request{}
var str = `{
    "version": 1,
    "componentName": "QC_SSLPOD",
    "eventId": 1298498081,
    "timestamp": 1553845083,
    "interface": {
        "interfaceName": "qcloud.sslpod.test",
        "para": {
            "a": {
                "b": 1,
                "c": "2",
                "d": true,
				"e": [
					"1",
					"2",
					"3"
				],
				"f": {
					"a": 1,
					"b": "2",
					"c": true
				}
			},
            "b": 10,
            "c": "2",
            "d": true,
			"e": [
				"1",
				"2",
				"3"
			],
			"f": {
				"a": 1,
				"b": "2",
				"c": true
			},
			"g": {
				"a": "1"
			},
			"h": [
				1,
				2
			]
        }
    }
}`

func init() {
	err := json.Unmarshal([]byte(str), req)
	if err != nil {
		panic(err)
	}
}

func TestGetBool(t *testing.T) {
	b := req.GetBool("a.d")
	assert.True(t, b)
	b = req.GetBool("d")
	assert.True(t, b)
}

func TestGetString(t *testing.T) {
	s := req.GetString("a.f.b")
	assert.NotEmpty(t, s)
	s = req.GetString("c")
	assert.NotEmpty(t, s)
}

func TestGetInt(t *testing.T) {
	i := req.GetInt("a.b")
	assert.NotEmpty(t, i)
	i = req.GetInt("b")
	assert.NotEmpty(t, i)
}

func TestGetIntSlice(t *testing.T) {
	is := req.GetIntSlice("a.b")
	assert.IsType(t, []int{}, is)
}

func TestGetInt64(t *testing.T) {
	var c int64 = 10
	i := req.GetInt64("a.b")
	assert.IsType(t, c, i)
	i = req.GetInt64("b")
	assert.IsType(t, c, i)
}

func TestGetFloat64(t *testing.T) {
	f := req.GetFloat64("b")
	assert.IsType(t, float64(1.0), f)
	f = req.GetFloat64("a.b")
	assert.IsType(t, float64(1.0), f)
}

func TestGetStringSlice(t *testing.T) {
	sli := req.GetStringSlice("e")
	t.Log(sli)
	assert.IsType(t, []string{"1"}, sli)
	sli = req.GetStringSlice("a.e")
	assert.IsType(t, []string{"1"}, sli)
}

func TestGetStringMap(t *testing.T) {
	m := req.GetStringMap("f")
	assert.IsType(t, map[string]interface{}{}, m)
}

func TestGetStringMapString(t *testing.T) {
	m := req.GetStringMapString("g")
	assert.IsType(t, map[string]string{}, m)
}
