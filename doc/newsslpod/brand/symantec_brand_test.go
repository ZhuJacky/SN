package brand

import (
	"context"
	"testing"
)

func TestQuerySymantec(t *testing.T) {
	var data = []struct {
		cn         string
		sn         string
		isSymantec bool
	}{
		{"www.trustasia.com", "1140BB2344209C7895FE26F5B73388C", true},    // 多个证书，先返回列表，再遍历详情
		{"web.e-chinalife.com", "0fb7cbc3d722290159e769d58bcc7241", true}, // 单个证书，直接详情
		{"*.51lianjin.com", "074a85b447e8ff82e6ca0215ef5cd605", true},
		{"www.digicert.com", "793EC89595DBA606D1FD9F7BE389802", false},
	}
	s := &SymantecBrandQuery{}
	s.Init(context.Background())
	for i := 0; i < len(data); i++ {
		t.Log("test ", data[i].cn)
		if s.QuerySymantec(data[i].cn, data[i].sn) != data[i].isSymantec {
			t.Fatal(data[i].cn, data[i].sn, data[i].isSymantec)
		}
	}
}
