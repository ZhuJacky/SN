// Package qcloud provides ...
package qcloud

import (
	"encoding/json"
	"testing"
)

func TestResponseGetInt(t *testing.T) {
	str := []byte(`{"version":1,"timestamp":0,"eventId":5577006791947779410,"componentName":"QC_MESSAGE","returnValue":2,"returnCode":20001,"returnMessage":"send msg not perm | caller is invalid","interface":"message.message.sendMsg2SubAccount","data":[]}`)

	resp := &Response{}
	err := json.Unmarshal(str, resp)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.ReturnValue)
}
