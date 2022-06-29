// Package mc provides ...
package mc

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"mysslee_qcloud/qcloud"
)

func HandleRequest(req *qcloud.Request) (*qcloud.Response, error) {
	resp := &qcloud.Response{
		Version:       1,
		Timestamp:     time.Now().Unix(),
		EventId:       rand.Int63(),
		ComponentName: "QC_ACCOUNT",
		ReturnCode:    0,
		Interface:     req.Interface.InterfaceName,
	}

	interfaceName := req.Interface.InterfaceName
	apiName := strings.Split(interfaceName, ".")[2]

	api, ok := responseData[apiName]
	if !ok {
		return resp, errors.New("not found api")
	}
	str, err := api(req.Interface.Para)
	if err != nil {
		return resp, err
	}
	resp.Data = json.RawMessage(str)
	return resp, nil
}

var responseData = map[string]func(map[string]interface{}) (string, error){
	"getAppIdByLoginUin":    GetAppIdByLoginUin,
	"getUserInfoByLoginUin": GetUserInfoByLoginUin,
	"sendMsg2SubAccount":    SendMsg2SubAccount,
	"sendMsg4Subscribe":     SendMsg4Subscribe,
	"sendMsg4Unsubscribe":   SendMsg4Unsubscribe,
}

func GetAppIdByLoginUin(para map[string]interface{}) (string, error) {
	uin, ok := para["loginUin"]
	if !ok {
		return "", errors.New("not found loginUin")
	}
	fmt.Println("uin=", uin)
	return "[1251000011]", nil
}

func GetUserInfoByLoginUin(para map[string]interface{}) (string, error) {
	uin, ok := para["loginUin"]
	if !ok {
		return "", errors.New("not found loginUin")
	}
	fmt.Println("uin=", uin)
	return `{
    "user_type": "creator",
    "uin": "2407912486",
    "owner_uin": 909619400,
    "type": "0",
    "name": "henry.chen",
    "mail": "henry.chen@trustasia.com",
    "cur_mail_pass": 1,
    "tel": "13276389503",
    "auth_method": 0
}`, nil
}

func SendMsg2SubAccount(para map[string]interface{}) (string, error) {
	uin, ok := para["ownerUin"]
	if !ok {
		return "", errors.New("not found ownerUin")
	}
	themeId, ok := para["themeId"]
	if !ok {
		return "", errors.New("not found themeId")
	}
	lang, ok := para["lang"]
	if !ok {
		return "", errors.New("not found lang")
	}
	channel, ok := para["sendChannel"]
	if !ok {
		return "", errors.New("not found sendChannel")
	}

	flag := int(channel.(float64))
	if flag&1 == 1 {
		fmt.Printf(">>>发送站内信 uin=%v themeId=%v lang=%v\n", uin, themeId, lang)
	}
	if flag&2 == 2 {
		fmt.Printf(">>>发送邮件 uin=%v themeId=%v lang=%v\n", uin, themeId, lang)
	}
	if flag&4 == 4 {
		fmt.Printf(">>>发送短信 uin=%v themeId=%v lang=%v\n", uin, themeId, lang)
	}
	if flag&8 == 8 {
		fmt.Printf(">>>发送微信 uin=%v themeId=%v lang=%v\n", uin, themeId, lang)
	}
	return `{"logId":292015}`, nil
}

func SendMsg4Subscribe(para map[string]interface{}) (string, error) {
	uin, ok := para["ownerUin"]
	if !ok {
		return "", errors.New("not found ownerUin")
	}
	themeId, ok := para["themeId"]
	if !ok {
		return "", errors.New("not found themeId")
	}
	fmt.Printf(">>>选择了 uin=%v themeId=%v\n", uin, themeId)
	return `{"logId":31}`, nil
}

func SendMsg4Unsubscribe(para map[string]interface{}) (string, error) {
	uin, ok := para["ownerUin"]
	if !ok {
		return "", errors.New("not found ownerUin")
	}
	themeId, ok := para["themeId"]
	if !ok {
		return "", errors.New("not found themeId")
	}
	lang, ok := para["lang"]
	if !ok {
		return "", errors.New("not found lang")
	}
	channel, ok := para["sendChannel"]
	if !ok {
		return "", errors.New("not found sendChannel")
	}
	receiver, ok := para["receiver"]
	if !ok {
		return "", errors.New("not found receiver")
	}

	flag, err := strconv.Atoi(channel.(string))
	if err != nil {
		return "", errors.New("invalid sendChannel")
	}
	if flag&1 == 1 {
		fmt.Printf(">>>发送站内信 uin=%v themeId=%v lang=%v receiver=%v\n", uin, themeId, lang, receiver)
	} else if flag&2 == 2 {
		fmt.Printf(">>>发送邮件 uin=%v themeId=%v lang=%v receiver=%v\n", uin, themeId, lang, receiver)
	} else if flag&4 == 4 {
		fmt.Printf(">>>发送短信 uin=%v themeId=%v lang=%v receiver=%v\n", uin, themeId, lang, receiver)
	} else if flag&8 == 8 {
		fmt.Printf(">>>发送微信 uin=%v themeId=%v lang=%v receiver=%v\n", uin, themeId, lang, receiver)
	}
	return `{"logId":31}`, nil
}
