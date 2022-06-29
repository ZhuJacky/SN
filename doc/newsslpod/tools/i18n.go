// Package main provides ...
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	gopath   = os.Getenv("GOPATH")
	source   = gopath + "/src/mysslee_qcloud/services/errcode.go"
	targetEN = gopath + "/src/mysslee_qcloud/conf/i18n/en-us.ini"
	targetCN = gopath + "/src/mysslee_qcloud/conf/i18n/zh-cn.ini"
	re       = regexp.MustCompile(`[\w\s_]+=[\s-\d]+//([\p{Han}_,:@，%\s\w]+&[\w\s,]+)`)
)

func main() {
	f, err := os.Open(source)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	enF, err := os.OpenFile(targetEN, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer enF.Close()
	zhF, err := os.OpenFile(targetCN, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer zhF.Close()
	reader := bufio.NewReader(f)
	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Println(err)
				break
			}
			if len(data) == 0 {
				break
			}
		}
		line := strings.TrimSpace(string(data))
		if re.MatchString(line) {
			desc := strings.Split(line, "=")[1]
			idx := strings.Index(desc, "//")
			code := strings.TrimSpace(desc[:idx])
			langs := strings.Split(desc[idx+2:], "&")
			if len(langs) > 0 {
				zhF.WriteString(fmt.Sprintf("%s = %s\n", code, strings.TrimSpace(langs[0])))
			}
			if len(langs) > 1 {
				enF.WriteString(fmt.Sprintf("%s = %s\n", code, strings.TrimSpace(langs[1])))
			}
		} else {
			// fmt.Println(re.MatchString(line), line)
		}
	}
}
