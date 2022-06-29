package myconn

import "testing"

func TestIPRequest(t *testing.T) {
	resp, err := HttpRequest(nil, "GET", "https://115.29.145.169/", "deepzz.com", 0, nil, nil)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("statusCode: %d, length: %v", resp.StatusCode, resp.Header)
		resp.Body.Close()
	}

	resp, err = HttpRequest(nil, "GET", "http://yryz.net/", "", 0, nil, nil)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("statusCode: %d, length: %v", resp.StatusCode, resp.Header)
		resp.Body.Close()
	}

	// 百度的实现如果带了:443, 会导致响应 405
	resp, err = HttpRequest(nil, "GET", "https://www.baidu.com:443/favicon.ico", "", 0, nil, nil)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("statusCode: %d, length: %v", resp.StatusCode, resp.Header)
		resp.Body.Close()
	}
}
