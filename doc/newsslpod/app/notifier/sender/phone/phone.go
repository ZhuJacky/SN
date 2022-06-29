// Package phone provides ...
package phone

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"mysslee_qcloud/config"
	"mysslee_qcloud/utils"
)

// 电话格式
type Tel struct {
	Mobile     string `json:"mobile"`
	Nationcode string `json:"nationcode"`
}

// 腾讯云短信
type SMS struct {
	Ext    string `json:"ext"`
	Extend string `json:"extend"`
	Msg    string `json:"msg"`
	Sig    string `json:"sig"`
	Tel    Tel    `json:"tel"`
	Time   int64  `json:"time"`
	Type   int    `json:"type"`
}

type SMSResp struct {
	Result int    `json:"result"`
	Errmsg string `json:"errmsg"`
	Ext    string `json:"ext"`
	Fee    int    `json:"fee"`
	Sid    string `json:"sid"`
}

// 短信
func SendSMS(msg, to, ext string) error {
	now := time.Now()

	rand := utils.RandomCode(6, false)
	url := fmt.Sprintf("%s/tlssmssvr/sendsms?sdkappid=%s&random=%s",
		config.Conf.Notifier.Phone.API, config.Conf.Notifier.Phone.User, rand)

	// signature
	plaintext := fmt.Sprintf("appkey=%s&random=%s&time=%d&mobile=%s",
		config.Conf.Notifier.Phone.Key, rand, now.Unix(), to)
	h := sha256.New()
	h.Write([]byte(plaintext))
	sig := fmt.Sprintf("%x", h.Sum(nil))

	sms := &SMS{
		Msg:  msg,
		Time: now.Unix(),
		Type: 0,
		Sig:  sig,
	}
	sms.Tel.Mobile = to
	sms.Tel.Nationcode = "86"
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := enc.Encode(sms)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	smsResp := new(SMSResp)
	err = dec.Decode(smsResp)
	if err != nil {
		return err
	}
	if smsResp.Result != 0 {
		return errors.New(smsResp.Errmsg)
	}
	return nil
}

// 腾讯云电话
type Call struct {
	PlayTimes  int    `json:"playtimes"`
	PromptFile string `json:"promptfile"`
	PromptType int    `json:"prompttype"`
	Sig        string `json:"sig"`
	Tel        Tel    `json:"tel"`
	Time       int64  `json:"time"`
}

type CallResp struct {
	Result int    `json:"result"`
	Errmsg string `json:"errmsg"`
	CallId string `json:"callid"`
	Ext    string `json:"ext"`
}

// 电话
func CallPhone(msg, to, ext string) error {
	now := time.Now()

	rand := utils.RandomCode(6, false)
	url := fmt.Sprintf("%s/tlsvoicesvr/sendvoiceprompt?sdkappid=%s&random=%s",
		config.Conf.Notifier.Phone.API, config.Conf.Notifier.Phone.User, rand)

	// signature
	plaintext := fmt.Sprintf("appkey=%s&random=%s&time=%d&mobile=%s",
		config.Conf.Notifier.Phone.Key, rand, now.Unix(), to)
	h := sha256.New()
	h.Write([]byte(plaintext))
	sig := fmt.Sprintf("%x", h.Sum(nil))

	call := &Call{
		PlayTimes:  2,
		PromptFile: msg,
		PromptType: 2,
		Sig:        sig,
		Time:       now.Unix(),
	}
	call.Tel.Mobile = to
	call.Tel.Nationcode = "86"
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := enc.Encode(call)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	callResp := new(CallResp)
	err = dec.Decode(callResp)
	if err != nil {
		return err
	}
	if callResp.Result != 0 {
		return errors.New(callResp.Errmsg)
	}
	return nil
}
