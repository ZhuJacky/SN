package ct

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	ct "github.com/google/certificate-transparency-go"
	"github.com/pkg/errors"
	//"log"
	//log "qiniupkg.com/x/log.v7"
)

//CT信息
type CTInfo struct {
	Logs      []*CTLogInfo `json:"logs"`
	Operators []*Operator  `json:"operators"`
}

type Operator struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

func (o *Operator) String() string {
	return fmt.Sprintf("id:%v -- name:%v", o.Id, o.Name)
}

type CTLogInfo struct {
	Description       string    `json:"description"`
	Key               string    `json:"key"`
	Url               string    `json:"url"`
	MaximumMergeDelay int       `json:"maximum_merge_delay"`
	OperatedBy        []int     `json:"operated_by"`
	FinalSth          *FinalSth `json:"final_sth"`
	DisqualifiedAt    int64     `json:"disqualified_at"`
	DnsApiEndpoint    string    `json:"dns_api_endpoint"`
}

func (c *CTLogInfo) String() string {
	return fmt.Sprintf("description:%v  key:%v  url:%v  MaximumMergeDelay:%v  DisqualifiedAt:%v  DnsApiEndpoint:%v ", c.Description, c.Key, c.Url, c.MaximumMergeDelay, c.DisqualifiedAt, c.DnsApiEndpoint)
}

type FinalSth struct {
	TreeSize          int    `json:"tree_size"`
	Timestamp         int    `json:"timestamp"`
	Sha256RootHash    string `json:"sha256_root_hash"`
	TreeHeadSignature string `json:"tree_head_signature"`
}

//获取并更新证书信息
func InitCT() {
	time.AfterFunc(3*time.Minute, initCT)
}

func initCT() {
	defer time.AfterFunc(24*time.Hour, InitCT)
	result, err := getCTLogsInfo()
	if err == nil {
		ctmap, _ := GetMapCTInfo(result)
		//如果变化超过了原始的1/2 则不进行处理
		//if math.Abs(float64(len(CTLogList)-len(ctmap))) > float64(len(ctmap)/2) {
		//	log.Error("CT信息更行超过了1/2")
		//	return
		//}

		if (len(CTLogList) - len(ctmap)) > len(CTLogList)/2 {
			//log.Errorf("CT信息比预置的减少了1/2, 变化太大，放弃自动更新")
			return
		}
		CTLogList = ctmap
	}
}

func getCTLogsInfo() (result *CTInfo, err error) {
	config := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "www.gstatic.com",
	}
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: config,
		},
	}
	req, err := http.NewRequest(http.MethodGet, "https://www.gstatic.com/ct/log_list/log_list.json", nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	infos := &CTInfo{}
	err = json.Unmarshal(data, &infos)
	if err != nil {
		return nil, err
	}
	return infos, nil
}

//通过ct信息中的公钥计算id
func calculateID(key string) (id string, err error) {
	pk := fmt.Sprintf("%s\n%s\n%s", "-----BEGIN PUBLIC KEY-----", key, "-----END PUBLIC KEY-----")
	_, pkb, _, err := ct.PublicKeyFromPEM([]byte(pk))
	if err != nil {
		return "", err
	}
	id = base64.StdEncoding.EncodeToString(pkb[:])
	return id, nil
}

//生成ct的map信息
func GetMapCTInfo(info *CTInfo) (cts map[string]*CTLogInfo, err error) {
	if info == nil {
		return nil, errors.New("没有CT信息")
	}
	cts = make(map[string]*CTLogInfo, len(info.Logs))

	for _, log := range info.Logs {
		id, err := calculateID(log.Key)
		if err != nil {
			continue
		}
		cts[id] = log
	}

	return cts, nil
}

// CT日志管理人员
var CTOperators = map[int]string{
	0: "Google",
	1: "Cloudflare",
	2: "DigiCert",
	3: "Certly",
	4: "Izenpe",
	5: "WoSign",
	6: "Venafi",
	7: "CNNIC",
	8: "StartCom",
	9: "Comodo CA Limited",
}

// 20180719更新
var CTLogList = map[string]*CTLogInfo{
	"pFASaQVaFVReYhGrN7wQP2KuVXakXksXFEU+GyIQaiU=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE0gBVBa3VR7QZu82V+ynXWD14JM3ORp37MtRxTmACJV5ZPtfUA7htQ2hofuigZQs+bnFZkje+qejxoyvk2Q1VaA==",
		Description:    "Google 'Argon2018' log",
		Url:            "ct.googleapis.com/logs/argon2018/",
		DnsApiEndpoint: "argon2018.ct.googleapis.com",
		OperatedBy:     []int{0},
	},

	"Y/Lbzeg7zCzPC3KEJ1drM6SNYXePvXWmOLHHaFRL2I0=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEI3MQm+HzXvaYa2mVlhB4zknbtAT8cSxakmBoJcBKGqGwYS0bhxSpuvABM1kdBTDpQhXnVdcq+LSiukXJRpGHVg==",
		Description:    "Google 'Argon2019' log",
		Url:            "ct.googleapis.com/logs/argon2019/",
		DnsApiEndpoint: "argon2019.ct.googleapis.com",
		OperatedBy:     []int{0},
	},

	"sh4FzIuizYogTodm+Su5iiUgZ2va+nDnsklTLe+LkF4=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6Tx2p1yKY4015NyIYvdrk36es0uAc1zA4PQ+TGRY+3ZjUTIYY9Wyu+3q/147JG4vNVKLtDWarZwVqGkg6lAYzA==",
		Description:    "Google 'Argon2020' log",
		Url:            "ct.googleapis.com/logs/argon2020/",
		DnsApiEndpoint: "argon2020.ct.googleapis.com",
		OperatedBy:     []int{0},
	},

	"9lyUL9F3MCIUVBgIMJRWjuNNExkzv98MLyALzE7xZOM=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETeBmZOrzZKo4xYktx9gI2chEce3cw/tbr5xkoQlmhB18aKfsxD+MnILgGNl0FOm0eYGilFVi85wLRIOhK8lxKw==",
		Description:    "Google 'Argon2021' log",
		Url:            "ct.googleapis.com/logs/argon2021/",
		DnsApiEndpoint: "argon2021.ct.googleapis.com",
		OperatedBy:     []int{0},
	},

	"aPaY+B9kgr46jO65KB1M/HFRXWeT1ETRCmesu09P+8Q=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE1/TMabLkDpCjiupacAlP7xNi0I1JYP8bQFAHDG1xhtolSY1l4QgNRzRrvSe8liE+NPWHdjGxfx3JhTsN9x8/6Q==",
		Description:    "Google 'Aviator' log",
		Url:            "ct.googleapis.com/aviator/",
		DnsApiEndpoint: "aviator.ct.googleapis.com",
		OperatedBy:     []int{0},
	},

	"KTxRllTIOWW6qlD8WAfUt2+/WHopctykwwz05UVH9Hg=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETtK8v7MICve56qTHHDhhBOuV4IlUaESxZryCfk9QbG9co/CqPvTsgPDbCpp6oFtyAHwlDhnvr7JijXRD9Cb2FA==",
		Description:    "Google 'Icarus' log",
		Url:            "ct.googleapis.com/icarus/",
		DnsApiEndpoint: "icarus.ct.googleapis.com",
		OperatedBy:     []int{0},
	},

	"pLkJkLQYWBSHuxOizGdwCjw1mAT5G9+443fNDsgN3BA=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEfahLEimAoz2t01p3uMziiLOl/fHTDM0YDOhBRuiBARsV4UvxG2LdNgoIGLrtCzWE0J5APC2em4JlvR8EEEFMoA==",
		Description:    "Google 'Pilot' log",
		Url:            "ct.googleapis.com/pilot/",
		DnsApiEndpoint: "pilot.ct.googleapis.com",
		OperatedBy:     []int{0},
	},

	"7ku9t3XOYLrhQmkfq+GeZqMPfl+wctiDAMR7iXqo/cs=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEIFsYyDzBi7MxCAC/oJBXK7dHjG+1aLCOkHjpoHPqTyghLpzA9BYbqvnV16mAw04vUjyYASVGJCUoI3ctBcJAeg==",
		Description:    "Google 'Rocketeer' log",
		Url:            "ct.googleapis.com/rocketeer/",
		DnsApiEndpoint: "rocketeer.ct.googleapis.com",
		OperatedBy:     []int{0},
	},

	"u9nfvB+KcbWTlCOXqpJ7RzhXlQqrUugakJZkNo4e0YU=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEEmyGDvYXsRJsNyXSrYc9DjHsIa2xzb4UR7ZxVoV6mrc9iZB7xjI6+NrOiwH+P/xxkRmOFG6Jel20q37hTh58rA==",
		Description:    "Google 'Skydiver' log",
		Url:            "ct.googleapis.com/skydiver/",
		DnsApiEndpoint: "skydiver.ct.googleapis.com",
		OperatedBy:     []int{0},
	},

	"23Sv7ssp7LH+yj5xbSzluaq7NveEcYPHXZ1PN7Yfv2Q=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEAsVpWvrH3Ke0VRaMg9ZQoQjb5g/xh1z3DDa6IuxY5DyPsk6brlvrUNXZzoIg0DcvFiAn2kd6xmu4Obk5XA/nRg==",
		Description:    "Cloudflare 'Nimbus2018' Log",
		Url:            "ct.cloudflare.com/logs/nimbus2018/",
		DnsApiEndpoint: "cloudflare-nimbus2018.ct.googleapis.com",
		OperatedBy:     []int{1},
	},

	"dH7agzGtMxCRIZzOJU9CcMK//V5CIAjGNzV55hB7zFY=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEkZHz1v5r8a9LmXSMegYZAg4UW+Ug56GtNfJTDNFZuubEJYgWf4FcC5D+ZkYwttXTDSo4OkanG9b3AI4swIQ28g==",
		Description:    "Cloudflare 'Nimbus2019' Log",
		Url:            "ct.cloudflare.com/logs/nimbus2019/",
		DnsApiEndpoint: "cloudflare-nimbus2019.ct.googleapis.com",
		OperatedBy:     []int{1},
	},

	"Xqdz+d9WwOe1Nkh90EngMnqRmgyEoRIShBh1loFxRVg=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE01EAhx4o0zPQrXTcYjgCt4MVFsT0Pwjzb1RwrM0lhWDlxAYPP6/gyMCXNkOn/7KFsjL7rwk78tHMpY8rXn8AYg==",
		Description:    "Cloudflare 'Nimbus2020' Log",
		Url:            "ct.cloudflare.com/logs/nimbus2020/",
		DnsApiEndpoint: "cloudflare-nimbus2020.ct.googleapis.com",
		OperatedBy:     []int{1},
	},

	"RJRlLrDuzq/EQAfYqP4owNrmgr7YyzG1P9MzlrW2gag=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAExpon7ipsqehIeU1bmpog9TFo4Pk8+9oN8OYHl1Q2JGVXnkVFnuuvPgSo2Ep+6vLffNLcmEbxOucz03sFiematg==",
		Description:    "Cloudflare 'Nimbus2021' Log",
		Url:            "ct.cloudflare.com/logs/nimbus2021/",
		DnsApiEndpoint: "cloudflare-nimbus2021.ct.googleapis.com",
		OperatedBy:     []int{1},
	},

	"VhQGmi/XwuzT9eG9RLI+x0Z2ubyZEVzA75SYVdaJ0N0=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEAkbFvhu7gkAW6MHSrBlpE1n4+HCFRkC5OLAjgqhkTH+/uzSfSl8ois8ZxAD2NgaTZe1M9akhYlrYkes4JECs6A==",
		Description:    "DigiCert Log Server",
		Url:            "ct1.digicert-ct.com/log/",
		DnsApiEndpoint: "digicert.ct.googleapis.com",
		OperatedBy:     []int{2},
	},

	"h3W/51l8+IxDmV+9827/Vo1HVjb/SrVgwbTq/16ggw8=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEzF05L2a4TH/BLgOhNKPoioYCrkoRxvcmajeb8Dj4XQmNY+gxa4Zmz3mzJTwe33i0qMVp+rfwgnliQ/bM/oFmhA==",
		Description:    "DigiCert Log Server 2",
		Url:            "ct2.digicert-ct.com/log/",
		DnsApiEndpoint: "digicert2.ct.googleapis.com",
		OperatedBy:     []int{2},
	},

	"wRZK4Kdy0tQ5LcgKwQdw1PDEm96ZGkhAwfoHUWT2M2A=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAESYlKFDLLFmA9JScaiaNnqlU8oWDytxIYMfswHy9Esg0aiX+WnP/yj4O0ViEHtLwbmOQeSWBGkIu9YK9CLeer+g==",
		Description:    "DigiCert Yeti2018 Log",
		Url:            "yeti2018.ct.digicert.com/log/",
		DnsApiEndpoint: "digicert-yeti2018.ct.googleapis.com",
		OperatedBy:     []int{2},
	},

	"4mlLribo6UAJ6IYbtjuD1D7n/nSI+6SPKJMBnd3x2/4=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEkZd/ow8X+FSVWAVSf8xzkFohcPph/x6pS1JHh7g1wnCZ5y/8Hk6jzJxs6t3YMAWz2CPd4VkCdxwKexGhcFxD9A==",
		Description:    "DigiCert Yeti2019 Log",
		Url:            "yeti2019.ct.digicert.com/log/",
		DnsApiEndpoint: "digicert-yeti2019.ct.googleapis.com",
		OperatedBy:     []int{2},
	},

	"8JWkWfIA0YJAEC0vk4iOrUv+HUfjmeHQNKawqKqOsnM=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEURAG+Zo0ac3n37ifZKUhBFEV6jfcCzGIRz3tsq8Ca9BP/5XUHy6ZiqsPaAEbVM0uI3Tm9U24RVBHR9JxDElPmg==",
		Description:    "DigiCert Yeti2020 Log",
		Url:            "yeti2020.ct.digicert.com/log/",
		DnsApiEndpoint: "digicert-yeti2020.ct.googleapis.com",
		OperatedBy:     []int{2},
	},

	"XNxDkv7mq0VEsV6a1FbmEDf71fpH3KFzlLJe5vbHDso=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6J4EbcpIAl1+AkSRsbhoY5oRTj3VoFfaf1DlQkfi7Rbe/HcjfVtrwN8jaC+tQDGjF+dqvKhWJAQ6Q6ev6q9Mew==",
		Description:    "DigiCert Yeti2021 Log",
		Url:            "yeti2021.ct.digicert.com/log/",
		DnsApiEndpoint: "digicert-yeti2021.ct.googleapis.com",
		OperatedBy:     []int{2},
	},

	"IkVFB1lVJFaWP6Ev8fdthuAjJmOtwEt/XcaDXG7iDwI=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEn/jYHd77W1G1+131td5mEbCdX/1v/KiYW5hPLcOROvv+xA8Nw2BDjB7y+RGyutD2vKXStp/5XIeiffzUfdYTJg==",
		Description:    "DigiCert Yeti2022 Log",
		Url:            "yeti2022.ct.digicert.com/log/",
		DnsApiEndpoint: "digicert-yeti2022.ct.googleapis.com",
		OperatedBy:     []int{2},
	},

	"3esdK3oNT6Ygi4GtgWhwfi6OnQHVXIiNPRHEzbbsvsw=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEluqsHEYMG1XcDfy1lCdGV0JwOmkY4r87xNuroPS2bMBTP01CEDPwWJePa75y9CrsHEKqAy8afig1dpkIPSEUhg==",
		Description:    "Symantec log",
		Url:            "ct.ws.symantec.com/",
		DnsApiEndpoint: "symantec.ct.googleapis.com",
		OperatedBy:     []int{2},
	},

	"vHjh38X2PGhGSTNNoQ+hXwl5aSAJwIG08/aRfz7ZuKU=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6pWeAv/u8TNtS4e8zf0ZF2L/lNPQWQc/Ai0ckP7IRzA78d0NuBEMXR2G3avTK0Zm+25ltzv9WWis36b4ztIYTQ==",
		Description:    "Symantec 'Vega' log",
		Url:            "vega.ws.symantec.com/",
		DnsApiEndpoint: "symantec-vega.ct.googleapis.com",
		OperatedBy:     []int{2},
	},

	"FZcEiNe5l6Bb61JRKt7o0ui0oxZSZBIan6v71fha2T8=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEowJkhCK7JewN47zCyYl93UXQ7uYVhY/Z5xcbE4Dq7bKFN61qxdglnfr0tPNuFiglN+qjN2Syxwv9UeXBBfQOtQ==",
		Description:    "Symantec 'Sirius' log",
		Url:            "sirius.ws.symantec.com/",
		DnsApiEndpoint: "symantec-sirius.ct.googleapis.com",
		OperatedBy:     []int{2},
	},

	"zbUXm3/BwEb+6jETaj+PAC5hgvr4iW/syLL1tatgSQA=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAECyPLhWKYYUgEc+tUXfPQB4wtGS2MNvXrjwFCCnyYJifBtd2Sk7Cu+Js9DNhMTh35FftHaHu6ZrclnNBKwmbbSA==",
		Description:    "Certly.IO log",
		Url:            "log.certly.io/",
		DnsApiEndpoint: "certly.ct.googleapis.com",
		OperatedBy:     []int{3},
	},

	"dGG0oJz7PUHXUVlXWy52SaRFqNJ3CbDMVkpkgrfrQaM=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEJ2Q5DC3cUBj4IQCiDu0s6j51up+TZAkAEcQRF6tczw90rLWXkJMAW7jr9yc92bIKgV8vDXU4lDeZHvYHduDuvg==",
		Description:    "Izenpe log",
		Url:            "ct.izenpe.com/",
		DnsApiEndpoint: "izenpe1.ct.googleapis.com",
		OperatedBy:     []int{4},
	},

	"QbLcLonmPOSvG6e7Kb9oxt7m+fHMBH4w3/rjs7olkmM=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEzBGIey1my66PTTBmJxklIpMhRrQvAdPG+SvVyLpzmwai8IoCnNBrRhgwhbrpJIsO0VtwKAx+8TpFf1rzgkJgMQ==",
		Description:    "WoSign log",
		Url:            "ctlog.wosign.com/",
		DnsApiEndpoint: "wosign1.ct.googleapis.com",
		OperatedBy:     []int{5},
	},

	"rDua7X+pZ0dXFZ5tfVdWcvnZgQCUHpve/+yhMTt1eC0=": &CTLogInfo{
		Key:            "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAolpIHxdSlTXLo1s6H1OCdpSj/4DyHDc8wLG9wVmLqy1lk9fz4ATVmm+/1iN2Nk8jmctUKK2MFUtlWXZBSpym97M7frGlSaQXUWyA3CqQUEuIJOmlEjKTBEiQAvpfDjCHjlV2Be4qTM6jamkJbiWtgnYPhJL6ONaGTiSPm7Byy57iaz/hbckldSOIoRhYBiMzeNoA0DiRZ9KmfSeXZ1rB8y8X5urSW+iBzf2SaOfzBvDpcoTuAaWx2DPazoOl28fP1hZ+kHUYvxbcMjttjauCFx+JII0dmuZNIwjfeG/GBb9frpSX219k1O4Wi6OEbHEr8at/XQ0y7gTikOxBn/s5wQIDAQAB",
		Description:    "Venafi log",
		Url:            "ctlog.api.venafi.com/",
		DnsApiEndpoint: "venafi.ct.googleapis.com",
		OperatedBy:     []int{6},
	},

	"AwGd8/2FppqOvR+sxtqbpz5Gl3T+d/V5/FoIuDKMHWs=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEjicnerZVCXTrbEuUhGW85BXx6lrYfA43zro/bAna5ymW00VQb94etBzSg4j/KS/Oqf/fNN51D8DMGA2ULvw3AQ==",
		Description:    "Venafi Gen2 CT log",
		Url:            "ctlog-gen2.api.venafi.com/",
		DnsApiEndpoint: "venafi2.ct.googleapis.com",
		OperatedBy:     []int{6},
	},

	"pXesnO11SN2PAltnokEInfhuD0duwgPC7L7bGF8oJjg=": &CTLogInfo{
		Key:            "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAv7UIYZopMgTTJWPp2IXhhuAf1l6a9zM7gBvntj5fLaFm9pVKhKYhVnno94XuXeN8EsDgiSIJIj66FpUGvai5samyetZhLocRuXhAiXXbDNyQ4KR51tVebtEq2zT0mT9liTtGwiksFQccyUsaVPhsHq9gJ2IKZdWauVA2Fm5x9h8B9xKn/L/2IaMpkIYtd967TNTP/dLPgixN1PLCLaypvurDGSVDsuWabA3FHKWL9z8wr7kBkbdpEhLlg2H+NAC+9nGKx+tQkuhZ/hWR65aX+CNUPy2OB9/u2rNPyDydb988LENXoUcMkQT0dU3aiYGkFAY0uZjD2vH97TM20xYtNQIDAQAB",
		Description:    "CNNIC CT log",
		Url:            "ctserver.cnnic.cn/",
		DnsApiEndpoint: "cnnic.ct.googleapis.com",
		OperatedBy:     []int{7},
	},

	"NLtq1sPfnAPuqKSZ/3iRSGydXlysktAfe/0bzhnbSO8=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAESPNZ8/YFGNPbsu1Gfs/IEbVXsajWTOaft0oaFIZDqUiwy1o/PErK38SCFFWa+PeOQFXc9NKv6nV0+05/YIYuUQ==",
		Description:    "StartCom log",
		Url:            "ct.startssl.com/",
		DnsApiEndpoint: "startcom1.ct.googleapis.com",
		OperatedBy:     []int{8},
	},

	"VYHUwhaQNgFK6gubVzxT8MDkOHhwJQgXL6OqHQcT0ww=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE8m/SiQ8/xfiHHqtls9m7FyOMBg4JVZY9CgiixXGz0akvKD6DEL8S0ERmFe9U4ZiA0M4kbT5nmuk3I85Sk4bagA==",
		Description:    "Comodo 'Sabre' CT log",
		Url:            "sabre.ct.comodo.com/",
		DnsApiEndpoint: "comodo-sabre.ct.googleapis.com",
		OperatedBy:     []int{9},
	},

	"b1N2rDHwMRnYmQCkURX/dxUcEdkCwQApBo2yCJo32RM=": &CTLogInfo{
		Key:            "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE7+R9dC4VFbbpuyOL+yy14ceAmEf7QGlo/EmtYU6DRzwat43f/3swtLr/L8ugFOOt1YU/RFmMjGCL17ixv66MZw==",
		Description:    "Comodo 'Mammoth' CT log",
		Url:            "mammoth.ct.comodo.com/",
		DnsApiEndpoint: "comodo-mammoth.ct.googleapis.com",
		OperatedBy:     []int{9},
	},
}
