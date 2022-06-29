// Package payment provides ...
package payment

import (
	"encoding/json"
	"testing"

	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/app/backend/db/redis"
	"mysslee_qcloud/qcloud"
)

func init() {
	db.Init()
	redis.Init()
}

var req = &qcloud.Request{
	Version:       1,
	ComponentName: "BillingRoute",
	SeqId:         "fbeed262-59f0-4ed3-8496-a7c8380c342a",
	SpanId:        "https://buy.qcloud.com;21382;2",
	EventId:       123,
	Timestamp:     12345678,
	Interface: qcloud.Interface{
		InterfaceName: "",
		Para: map[string]interface{}{
			"appId":         float64(2223),
			"uin":           "2223",
			"operateUin":    "2223",
			"type":          "sslpod",
			"region":        float64(1),
			"zoneId":        float64(0),
			"payMode":       float64(1),
			"projectId":     float64(0),
			"tranId":        "12312312312319",
			"flowId":        float64(8),
			"resourceId":    "plan-hhBAy944093",
			"autoRenewFlag": float64(1),
			"resourceIds": []string{
				"plan-hhBAy944093",
			},
			"pageNo":   float64(0),
			"pageSize": float64(10),
			"goodsDetail": map[string]interface{}{
				"timeSpan":      float64(2),
				"timeUnit":      "y",
				"goodsNum":      float64(1),
				"autoRenewFlag": float64(0),
				"pid":           float64(15959),
				"oldConfig": map[string]interface{}{
					"pid": float64(15959),
				},
				"newConfig": map[string]interface{}{
					"pid": float64(15960),
				},
				"curDeadline": "2021-04-18 09:23:05",
			},
		},
	},
}

func TestCheckCreate(t *testing.T) {
	result, errCode := HandleCheckCreate(req)
	if errCode != "" {
		t.Fatal(errCode)
	}
	t.Log(result)
}

func TestCreateResource(t *testing.T) {
	result, errCode := HandleCreateResource(req)
	if errCode != "" {
		t.Fatal(errCode)
	}
	t.Log(result)
}

func TestQueryFlow(t *testing.T) {
	result, errCode := HandleQueryFlow(req)
	if errCode != "" {
		t.Fatal(errCode)
	}
	t.Log(result)
}

func TestCheckModify(t *testing.T) {
	result, errCode := HandleCheckModify(req)
	if errCode != "" {
		t.Fatal(errCode)
	}
	t.Log(result)
}

func TestModifyResource(t *testing.T) {
	result, errCode := HandleModifyResource(req)
	if errCode != "" {
		t.Fatal(errCode)
	}
	t.Log(result)
}

func TestCheckRenew(t *testing.T) {
	result, errCode := HandleCheckRenew(req)
	if errCode != "" {
		t.Fatal(errCode)
	}
	t.Log(result)
}

func TestRenewResource(t *testing.T) {
	result, errCode := HandleRenewResource(req)
	if errCode != "" {
		t.Fatal(errCode)
	}
	t.Log(result)
}

func TestSetRenewFlag(t *testing.T) {
	result, errCode := HandleSetRenewFlag(req)
	if errCode != "" {
		t.Fatal(errCode)
	}
	t.Log(result)
}

func TestGetAllAppIds(t *testing.T) {
	result, errCode := HandleGetAllAppIds(req)
	if errCode != "" {
		t.Fatal(errCode)
	}
	t.Log(result)
}

func TestGetUserResource(t *testing.T) {
	result, errCode := HandleGetUserResource(req)
	if errCode != "" {
		t.Fatal(errCode)
	}
	t.Log(result)
}

func TestQueryResources(t *testing.T) {
	data := []byte(`{"eventId":5538,"componentName":"BillingRoute","seqId":"unknown","spanId":"","interface":{"interfaceName":"qcloud.sslpod.queryResources","para":{"resourceIds":["plan-NC0VTc44057"],"region":1,"appId":251007812,"type":"sslpod","uin":"524602999"}}}`)

	req := new(qcloud.Request)
	err := json.Unmarshal(data, req)
	if err != nil {
		t.Fatal(err)
	}

	result, errCode := HandleQueryResources(req)
	if errCode != "" {
		t.Fatal(errCode)
	}
	t.Log(result)
}

func TestIsolateResource(t *testing.T) {

}

func TestDestroyResource(t *testing.T) {

}

func TestQueryDeadlineList(t *testing.T) {

}
