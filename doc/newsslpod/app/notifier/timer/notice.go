// Package timer provides ...
package timer

// 注册激活邮件
import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"mysslee_qcloud/config"
	"mysslee_qcloud/model"
	"mysslee_qcloud/qcloud"

	"github.com/sirupsen/logrus"
)

// TplParam 通知模板参数
type TplParam struct {
	TitleTpl  []string `json:"titleTpl"`
	SiteTpl   []string `json:"siteTpl"`
	EmailTpl  []string `json:"emailTpl"`
	SMSTpl    []string `json:"smsTpl"`
	WechatTpl []string `json:"wechatTpl"`
}

func sendMsg2SubAccountToQCloud(msg *model.NoticeMsg) (*qcloud.Response, error) {
	// Get TPL params
	params := make(map[string]string)
	err := json.Unmarshal([]byte(msg.Msg), &params)
	if err != nil {
		return nil, err
	}
	param := []string{msg.Uin, params["Nickname"], params["Domain"], params["Port"], params["Time"], params["TrustStatus"]}
	tp := &TplParam{
		TitleTpl:  []string{"账号ID", "昵称", "域名", "端口", "时间", "信任状态"},
		SiteTpl:   param,
		EmailTpl:  param,
		SMSTpl:    param,
		WechatTpl: param,
	}

	req := &qcloud.Request{
		Version:       1,
		Caller:        config.Conf.QCloud.Caller,
		ComponentName: "MC",
		EventId:       int64(rand.Int31()),
		Interface: qcloud.Interface{
			InterfaceName: "message.message.sendMsg4Subscribe",
			Para: map[string]interface{}{
				"ownerUin":  msg.Uin,
				"themeId":   config.Conf.QCloud.ThemeIds[0],
				"tplParams": tp,
				// "sendChannel": (^msg.NoticeAlready) & msg.Type,
				"env": config.Conf.QCloud.Env,
			},
		},
	}
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	// request
	response, err := http.Post(config.Conf.QCloud.NotifyGateway, "application/json", bytes.NewReader(reqData))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, errors.New(response.Status)
	}
	// read data
	respData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	// log
	logrus.Infof("%v %s request body=%s response body=%s", time.Now().Sub(now), "sendMsg4Subscribe", reqData, respData)

	resp := &qcloud.Response{}
	// Unmarshal
	err = json.Unmarshal(respData, resp)
	if err != nil {
		return nil, err
	}
	if resp.ReturnValue != 0 {
		return nil, errors.New(resp.ReturnMessage)
	}
	return resp, nil
}
