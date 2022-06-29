package core

import (
	"context"
	"fmt"
	"os"
	"testing"

	"mysslee_qcloud/brand"
	"mysslee_qcloud/config"
	"mysslee_qcloud/core/myconn"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Chdir("../../app/reportapp")

	config.ReportAppInit()
	brand.Init(brand.BrandDB{
		Driver:          config.AppConf.DB.Driver,
		Source:          config.AppConf.DB.DataSourceName,
		TrustCAConfPath: config.TrustCAConfPath,
		TrustCALockPath: config.TrustCALockPath,
	})
}

func TestCheckCAA(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		domain string
		result bool
	}{
		{"hsulei.com", false},
		{"google.com", true},
		{"www.symantec.com", true},
	}

	for _, t := range tests {
		result, _ := CheckCAA(context.Background(), t.domain)
		assert.Equal(t.result, result, t.domain)
	}
}

//func TestCheckSSL2CertInfo(t *testing.T) {
//	assert := assert.New(t)
//	tests := []struct {
//		domain     string
//		ip         string
//		port       string
//		serverType myconn.ServerType
//		mailDirect bool
//		result     bool
//	}{
//		{"check.hsulei.com", "54.223.248.88", "443", myconn.Web, true, true},
//	}
//
//	for _, t := range tests {
//		_, err := CheckSSL2CertInfo(context.Background(), &myconn.CheckParams{Domain: t.domain, Ip: t.ip, Port: t.port, ServerType: t.serverType, MailDirect: t.mailDirect}, true)
//		if t.result {
//			assert.Nil(err, t.domain)
//		} else {
//			assert.NotNil(err, t.domain)
//		}
//
//	}
//
//	//_, err := CheckSSL2CertInfo(context.Background(), "check.hsulei.com", "54.223.248.88", "443", myconn.SMTP, true)
//	//assert.Nil(err)
//}

func TestGetMultipleCertInfo(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		domain     string
		ip         string
		port       string
		serverType myconn.ServerType
		mailDirect bool
		certsCount int
	}{
		//{"sc1.ecscloud.com", "54.222.180.204", "443"},
		//{"www.trustasia.com", "54.223.64.100", "443", myconn.Web, false, 3},
		//{"self-signed.badssl.com", "104.154.89.105", "443", myconn.Web, false, 2},
		//{"hsulei.com", "123.57.46.95", "443", myconn.Web, false, 2},
		//{"dev.wevalid.com", "183.131.76.65", "443", myconn.Web, false, 1},
		//{"ssl3.badssl.cn", "54.222.223.243", "443", myconn.Web, false, 1},
		//{"hsulei.com", "123.57.46.95", "443", myconn.Web, false, 2},
		//{"www.jiarener.com", "150.138.216.175", "443", myconn.Web, false, 2},
		//{"smtp.mail.outlook.com", "207.46.163.170", "587", myconn.SMTP, false, 2},
		//{"www.icbc-axa.com", "180.169.81.4", "443", myconn.Web, false, 1},
		// {"middle.badssl.cn", "54.222.223.243", "443", myconn.Web, false, 1},
		{"sm.ds.gsma.com", "91.240.72.81", "443", myconn.Web, false, 1},
	}

	for _, t := range tests {
		certs, err := GetMultipleCertInfo(context.Background(), &myconn.CheckParams{Domain: t.domain, Ip: t.ip, Port: t.port, ServerType: t.serverType, MailDirect: t.mailDirect})
		assert.Nil(err)

		for _, v := range certs[0].Cert.CertsInfo {
			fmt.Printf("%+v \n\n", v)
		}

		if err == nil {
			assert.Equal(t.certsCount, len(certs), t.domain)
		}
	}
}

func TestGetTLS13Certificate(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		domain     string
		ip         string
		port       string
		serverType myconn.ServerType
		mailDirect bool
		haveErr    bool
	}{
		{"hsulei.com", "123.57.46.95", "443", myconn.Web, false, false},
	}

	for _, t := range tests {
		_, err := GetTLS13Certificate(context.Background(), &myconn.CheckParams{Domain: t.domain, Ip: t.ip, Port: t.port, ServerType: t.serverType, MailDirect: t.mailDirect})
		assert.Nil(err)

	}
}

func TestCheckCertInfo(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		domain     string
		ip         string
		port       string
		serverType myconn.ServerType
		mailDirect bool
		certsCount int
	}{
		{"sm.ds.gsma.com", "91.240.72.81", "443", myconn.Web, false, 1},
	}
	for _, t := range tests {
		certs, err := CheckCertInfo(context.Background(),
			&myconn.CheckParams{
				Domain:     t.domain,
				Ip:         t.ip,
				Port:       t.port,
				ServerType: t.serverType,
				MailDirect: t.mailDirect,
			}, eccCheck, true, false)

		assert.Nil(err)

		for _, v := range certs.ServerCertificates {
			fmt.Printf("common name: %+v \n\n", v.Subject.CommonName)
			fmt.Printf("is ca: %+v \n\n", v.IsCA)
			fmt.Printf("self sign: %+v", v.CheckSignatureFrom(v) == nil)
		}

		fmt.Println(certs.ServerCertificates[0].CheckSignatureFrom(certs.ServerCertificates[1]))
	}
}
