package crl

import (
	"context"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"mysslee_qcloud/utils"

	log "github.com/sirupsen/logrus"
)

//crl 所对应的asn1序号
var crlAsn1 = asn1.ObjectIdentifier{2, 5, 29, 31}

//CheckCRL 证书吊销检测
func CheckCRL(ctx context.Context, cert *x509.Certificate) (revoked bool, err error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	fingerPrint := sha1.Sum(cert.Raw)
	var crlAddress []string

	crlAddress = cert.CRLDistributionPoints
	//找到CRL扩展

	//没有该延展
	if len(crlAddress) == 0 {
		log.WithFields(log.Fields{
			"hash":  fingerPrint,
			"line":  utils.ShowCallerMessage(1),
			"event": "查询crl",
		}).Warnf("hash:%v 没有crl字段", fmt.Sprintf("%x", fingerPrint))
		return false, errors.New("没有crl字段")
	}
	//先找本地缓存
	filePath := "cache/crls/" + slashToUnderline(crlAddress[0]) //?文件路径如何处理
	exist := fileExist(filePath)
	//如果不存在，直接去服务器获取，然后保存
	if !exist {
		crlData, raw, err := getCrlData(ctx, crlAddress[0])
		if err != nil {
			return false, err
		}
		//缓存该文件
		saveCache(filePath, raw, false)
		return checkRevoke(cert, crlData), nil
	}

	//读本地缓存
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false, err
	}
	crlData, err := x509.ParseCRL(data)
	if err != nil {
		return false, err
	}
	if crlData.TBSCertList.NextUpdate.Local().Before(time.Now().Local()) { //缓存的时间小于当前时间
		//从新获取并缓存
		crlData, data, err = getCrlData(ctx, crlAddress[0])
		if err != nil {
			return false, err
		}
		defer saveCache(filePath, data, true)
	}
	return checkRevoke(cert, crlData), nil
}

func saveCache(filePath string, data []byte, cover bool) {
	if !cover {
		err := os.MkdirAll("cache/crls/", 0666)
		if err != nil {
			log.WithFields(log.Fields{
				"line":  utils.ShowCallerMessage(1),
				"event": "创建cache/crls/ 文件夹错误",
			}).Warnf("创建cache/crls/ 文件夹错误: %v", err)
			return
		}
		_, err = os.Create(filePath)
		if err != nil {
			log.WithFields(log.Fields{
				"line":  utils.ShowCallerMessage(1),
				"event": "创建crl缓存文件",
			}).Warnf("缓存文件创建错误 :%v", err)
			return
		}
	}
	ioutil.WriteFile(filePath, data, 066)
}

//getCrlData http方式获取CRL数据
func getCrlData(ctx context.Context, crlAddress string) (*pkix.CertificateList, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, crlAddress, nil)
	if err != nil {
		return nil, nil, err
	}
	req = req.WithContext(ctx)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	crlData, err := x509.ParseCRL(data)
	if err != nil {
		return nil, nil, err
	}
	return crlData, data, nil

}

//checkRevoke 检测是否吊销
func checkRevoke(cert *x509.Certificate, crlData *pkix.CertificateList) bool {
	for _, revokeCertificat := range crlData.TBSCertList.RevokedCertificates {
		if revokeCertificat.SerialNumber == cert.SerialNumber {
			return true
		}
	}
	return false
}

func fileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil || os.IsExist(err)
}

//斜杠转换成下滑线
func slashToUnderline(content string) string {
	return strings.Replace(content, "/", "_", -1)
}

//下划线转换成斜杠
func underlineToSlash(content string) string {
	return strings.Replace(content, "_", "/", -1)
}
