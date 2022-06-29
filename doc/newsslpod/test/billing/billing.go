// Package billing provides ...
package billing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"text/template"
	"time"

	"mysslee_qcloud/config"
	"mysslee_qcloud/qcloud"
	"mysslee_qcloud/utils"

	"github.com/gin-gonic/gin"
)

var info = checkInfo{
	Type:          "预付费",
	TimeUnit:      "y",
	AutoRenewFlag: 0,
}

// 查询价格
func HandleQueryPrice(c *gin.Context, req qcloud.RequestBack) ([]byte, string) {
	if req.GetString("Action") != "SwitchParameterPlan" {
		return nil, qcloud.ErrInvalidAction
	}

	info.AppId = req.GetString("AppId")
	info.Uin = req.GetString("Uin")
	info.OperateUin = info.Uin
	info.GoodsCode = "plan"
	info.Region = 1
	info.ZoneId = 0
	info.PayMode = 1
	info.ProjectId = 0

	num := req.GetInt("GoodsNum")
	info.GoodsNum = num
	pid := req.GetInt("Pid")
	switch pid {
	case 15958:
		info.Pid = 15958
		info.UnitPrice = 0
		info.Name = "企业基础版"
		info.Describe = "短信：50<br/>邮件50"
	case 15959:
		info.Pid = 15959
		info.UnitPrice = 4999
		info.Name = "企业专业版"
		info.Describe = "短信：100<br/>邮件100"
	case 15960:
		info.Pid = 15960
		info.UnitPrice = 10999
		info.Name = "企业旗舰版"
		info.Describe = "短信：200<br/>邮件200"
	default:
		return nil, qcloud.ErrResourceNotFound
	}
	info.TimeSpan = req.GetInt("TimeSpan")
	if info.TimeSpan < 1 {
		return nil, qcloud.ErrInvalidParameterValue
	}
	info.Price = info.UnitPrice * info.TimeSpan
	price := fmt.Sprintf(`{"Price": %d}`, info.Price)
	resp := qcloud.ResponseBack{
		Response: json.RawMessage(price),
	}
	data, _ := json.Marshal(resp)

	return data, ""
}

// 检测页面
func HandleOrderCheck(c *gin.Context) {
	// TODO chekc paramter
	req := qcloud.Request{
		Version:       1,
		ComponentName: "BillingRoute",
		SeqId:         utils.RandomCode(32, false),
		SpanId:        "https://buy.qcloud.com;21382;2",
		EventId:       rand.Int63(),
		Timestamp:     time.Now().Unix(),
		Interface: qcloud.Interface{
			InterfaceName: "qcloud.sslpod.checkCreate",
			Para: map[string]interface{}{
				"appId":       info.AppId,
				"uin":         info.Uin,
				"operateUin":  info.OperateUin,
				"type":        info.GoodsCode,
				"region":      info.Region,
				"zoneId":      info.ZoneId,
				"payMode":     info.PayMode,
				"projectId":   info.ProjectId,
				"goodsDetail": info,
			},
		},
	}
	data, err := json.Marshal(req)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	response, err := http.Post(config.Conf.QCloud.BillingGateway, "application/json", bytes.NewReader(data))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer response.Body.Close()

	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	resp := &qcloud.Response{}
	err = json.Unmarshal(data, resp)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if resp.ReturnValue != 0 {
		c.String(http.StatusBadRequest, resp.ReturnMessage)
		return
	}
	status := resp.GetInt("status")
	if status != 0 {
		c.String(http.StatusBadRequest, "校验参数失败")
		return
	}

	t, err := template.New("check").Parse(tpl)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	t.Execute(c.Writer, info)
}

const tpl = `
<table border="1">
  <tr>
    <th>产品名称<//th>
    <th>配置信息<//th>
    <th>单价<//th>
    <th>数量<//th>
    <th>付款方式<//th>
    <th>购买时长<//th>
    <th>费用<//th>
  </tr>
  <tr>
    <td>{{.Name}}</td>
    <td>{{.Describe}}</td>
    <td>{{.UnitPrice}}</td>
    <td>{{.GoodsNum}}</td>
    <td>{{.Type}}</td>
    <td>{{.TimeSpan}}</td>
    <td>{{.Price}}</td>
  </tr>
</table>
`

type checkInfo struct {
	Name          string `json:"-"`
	Describe      string `json:"-"`
	UnitPrice     int    `json:"-"`
	Type          string `json:"-"`
	Price         int    `json:"-"`
	AppId         string `json:"-"`
	Uin           string `json:"-"`
	OperateUin    string `json:"-"`
	GoodsCode     string `json:"-"`
	Region        int    `json:"-"`
	ZoneId        int    `json:"-"`
	PayMode       int    `json:"-"`
	ProjectId     int    `json:"-"`
	Pid           int    `json:"pid"`
	GoodsNum      int    `json:"goodsNum"`
	TimeUnit      string `json:"timeUnit"`
	TimeSpan      int    `json:"timeSpan"`
	AutoRenewFlag int    `json:"autoRenewFlag"`
}

type baseInfo struct {
	AppId      string
	Uin        string
	OperateUin string
	Type       string
}
