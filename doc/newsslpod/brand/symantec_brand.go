package brand

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/json-iterator/go"
	"golang.org/x/net/publicsuffix"
)

// SymantecBrandQuery 通过证书在线查询Symantec品牌
type SymantecBrandQuery struct {
	client    *http.Client
	csrfToken string
	ctx       context.Context
}

func (s *SymantecBrandQuery) Init(ctx context.Context) {
	s.ctx = ctx
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	s.client = &http.Client{
		Jar: jar,
	}
	s.csrfToken = s.getCSRFToken()
	log.Println("symantec csrfToken:", s.csrfToken)
}

var regexCSRFToken = regexp.MustCompile(` csrfToken = "(.+)"`)

func (s *SymantecBrandQuery) getCSRFToken() string {

	r, _ := http.NewRequest(http.MethodGet, "https://trustcenter.websecurity.symantec.com/process/trust/search", nil)
	if s.ctx != nil {
		r = r.WithContext(s.ctx)
	}
	resp, err := s.client.Do(r)
	if err != nil {
		log.Println("getCSRFToken err:", err)
		return ""
	}

	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("getCSRFToken read err:", err)
		return ""
	}

	matchs := regexCSRFToken.FindStringSubmatch(string(buf))
	if len(matchs) > 1 {
		return matchs[1]
	}
	return ""
}

// 从Symantec处查询有效的证书，
func (s *SymantecBrandQuery) QuerySymantec(commonName, sn string) (exists bool) {
	if commonName == "" || sn == "" {
		return false
	}
	// data := fmt.Sprintf("commonName=%s&orderNumber=&all=false&valid=true&pending=false&revoked=false&expired=false&isFormValid=true&csrfToken=%s",
	// commonName, csrfToken)

	data := url.Values{}
	data.Set("commonName", commonName)
	data.Set("orderNumber", "")
	data.Set("all", "false")
	data.Set("valid", "true")
	data.Set("pending", "false")
	data.Set("revoked", "false")
	data.Set("expired", "false")
	data.Set("isFormValid", "true")
	data.Set("csrfToken", s.csrfToken)

	log.Println("searchResult")
	r, _ := http.NewRequest(http.MethodPost, "https://trustcenter.websecurity.symantec.com/process/trust/searchService/searchResult",
		strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	// r.Header.Add("Referer", "https://trustcenter.websecurity.symantec.com/process/trust/search")
	// r.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36")
	// r.Header.Set("X-Requested-With", "XMLHttpRequest")
	// r.Header.Set("csrftoken", csrfToken)
	if s.ctx != nil {
		r = r.WithContext(s.ctx)
	}

	resp, err := s.client.Do(r)
	if err != nil {
		log.Println(err)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	// fmt.Println("body", resp.Status, resp.Header, string(body))
	serialNumber := jsoniter.Get(body, "orderDetails", "serialNumber").ToString()
	if serialNumber != "" {
		// 我们算的序列号前面没有补0，需要从后比较
		if strings.LastIndex(serialNumber, strings.ToLower(sn)) != -1 {
			return true
		}
		return false
	}

	for i := 0; i < 10; i++ {
		any := jsoniter.Get(body, "orderInfoList", i)
		if any.LastError() != nil {
			break
		}
		// log.Println(any.Get("issuerSerialDigest"), any.Get("orderNumber"))
		serialNumber := s.queryDetail(any.Get("issuerSerialDigest").ToString(), any.Get("orderNumber").ToString())
		log.Println(serialNumber)
		// 我们算的序列号前面没有补0，需要从后比较
		if strings.LastIndex(serialNumber, strings.ToLower(sn)) != -1 {
			return true
		}
	}

	return false
}

func (s *SymantecBrandQuery) queryDetail(issuerSerialDigest, orderNumber string) (serialNumber string) {
	data := url.Values{}
	data.Set("issuerSerialDigest", issuerSerialDigest)
	data.Set("orderNumber", orderNumber)
	data.Set("csrfToken", s.csrfToken)

	log.Println("searchDetails")
	r, _ := http.NewRequest(http.MethodPost, "https://trustcenter.websecurity.symantec.com/process/trust/searchService/searchDetails",
		strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	if s.ctx != nil {
		r = r.WithContext(s.ctx)
	}

	resp, err := s.client.Do(r)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	// fmt.Println("body", resp.Status, resp.Header, string(body))
	return jsoniter.Get(body, "orderDetails", "serialNumber").ToString()
}
