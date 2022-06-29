package ct

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCTLogsInfo(t *testing.T) {
	assert := assert.New(t)
	result, err := getCTLogsInfo()
	assert.Nil(err)
	res, err := GetMapCTInfo(result)
	assert.Nil(err)
	fmt.Printf("%+v", res)
}

func TestCalculateID(t *testing.T) {
	assert := assert.New(t)
	key := `MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE1/TMabLkDpCjiupacAlP7xNi0I1JYP8bQFAHDG1xhtolSY1l4QgNRzRrvSe8liE+NPWHdjGxfx3JhTsN9x8/6Q==`
	id, err := calculateID(key)
	assert.Nil(err)
	fmt.Printf("===\n%s\n===", id)
}

func TestGenLogList(t *testing.T) {
	assert := assert.New(t)
	result, err := getCTLogsInfo()
	assert.Nil(err)
	str := `// Package cert 该文件由TestGenLogList函数自动生成，请勿手动更改
package cert

var CTLogList = map[string]*CTLogInfo{`
	for _, v := range result.Logs {
		id, err := calculateID(v.Key)
		if err != nil {
			panic(err)
		}
		str = str + `
	"` + id + `": &CTLogInfo{
		Key: 			"` + v.Key + `",
		Description: 	"` + v.Description + `",
		Url: 			"` + v.Url + `",
		DnsApiEndpoint: "` + v.DnsApiEndpoint + `",
	},
		`
	}
	str = str + `
}`

	ioutil.WriteFile("./ct_log_list.go", []byte(str), 0777)
}
