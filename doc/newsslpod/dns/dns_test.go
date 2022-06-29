// Package dns provides ...
package dns

import (
	"context"
	"fmt"
	"testing"
	"time"

	"golang.org/x/net/idna"
)

var testDomains = []string{
	// "local.deepzz.com",
	// "baidu.com",
	"google.com",
	// "twitter.com",
	// "facebook.com",
	// "youtube.com",
	// "亚洲诚信.com",
	// "example.invalid",
	// "----begin certificate request---- miicydccazfgvafasdfsdfasdfasdfasdf",
}

var blackList = map[string]bool{
	"google.com":   true,
	"twitter.com":  true,
	"facebook.com": true,
	"youtube.com":  true,
}

func init() {
	DefaultClient = NewDNSClient(&ClientOpt{
		Net:         "udp",
		ReadTimeout: time.Second * 2,
		MaxTries:    3,
		QueryAAAA:   false,
		Servers: []string{
			"54.223.185.170:53",
			// "218.253.193.60:53",
			// "52.2.29.193:8053",
			// "223.5.5.5",
			// "218.253.193.60:53",
			// "52.2.29.193:8053",
			// "223.5.5.5",
		},
	}, nil)
	// DefaultClient.servers[1].LocationCode = "852"
}

func TestLookupHostEDNS(t *testing.T) {
	testDomains := []string{
		"54.223.185.170",
		"52.2.29.193",
	}

	for _, v := range testDomains {
		ctx := context.WithValue(context.Background(), "dns_eaddr", v)

		ips, err := LookupHost(ctx, "www.alibaba.com")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(ips)
	}
}

func TestLookupHost(t *testing.T) {
	for i, d := range testDomains {
		d, err := idna.ToASCII(d)
		if err != nil {
			t.Log(d, err)
			continue
		}
		ctx := context.Background()
		if blackList[d] {
		}
		ips, err := LookupHost(ctx, d)
		if i == 3 && err != nil {
			t.Log(d, err)
			continue
		} else if err != nil {
			t.Error(d, err)
			continue
		}
		t.Log(d, ips)
		time.Sleep(time.Second)
	}
}

func TestLookupIP(t *testing.T) {
	for i, d := range testDomains {
		ips, err := LookupIP(context.Background(), d)
		if i == 3 && err != nil {
			t.Log(d, err)
			continue
		} else if err != nil {
			t.Error(d, err)
			continue
		}
		t.Log(d, ips)
	}
}

func TestLookupMX(t *testing.T) {
	for _, d := range testDomains {
		records, err := LookupMX(context.Background(), d)
		if err != nil {
			t.Error(d, err)
			continue
		}
		for _, v := range records {
			t.Log(d, v.Mx, v.Preference)
		}
	}
}

func TestLookupTXT(t *testing.T) {
	for _, d := range testDomains {
		txt, err := LookupTXT(context.Background(), d)
		if err != nil {
			t.Error(d, err)
			continue
		}
		t.Log(d, txt)
	}
}

func TestLookupCAA(t *testing.T) {
	var testDomains = []string{
		// "google.com",
		// "www.dnsimple.com",
		// "support.dnsimple.com",
		// "hello.2kui.cn",
		// "st.deepzz.com",
		// "deepzz.com",
		// "laifengmeishi.com",
		// "cloudcard.lol",
		"www.puzhenyi.net",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	for _, d := range testDomains {
		caa, err := LookupCAA(ctx, d)
		if err != nil {
			t.Error(d, err)
			continue
		}
		t.Log(caa)
	}
}

func TestLookupCNAME(t *testing.T) {
	var testDomains = []string{
		"kpawmpskbl.freessl.org",
	}
	for _, d := range testDomains {
		cname, err := LookupCNAME(context.Background(), d)
		if err != nil {
			t.Error(d, err)
			continue
		}
		t.Log(d, cname)
	}
}

func TestLookupNS(t *testing.T) {
	var testDomains = []string{
		"baidu.com",
	}

	for _, d := range testDomains {
		ns, err := LookupNS(context.Background(), d)
		if err != nil {
			t.Error(d, err)
			continue
		}
		t.Log(d, ns)
	}
}

func BenchmarkLookupHost(b *testing.B) {
	domain := "baidu.com"
	b.N = 10000
	for i := 0; i < b.N; i++ {
		ips, err := LookupHost(context.Background(), domain)
		if err != nil {
			b.Error(i, err)
			continue
		}
		fmt.Println(i, ips)
	}
}

var testIPs = []string{
	"116.1.2.3",
	"2001:db8::68",
	"127.0.0.1",
	"192.168.99.100",
	"172.24.58.252",
}

func TestIsPrivateIP(t *testing.T) {
	for _, v := range testIPs {
		ok := IsPrivateIP(v)
		t.Log(v, ok)
	}
}
