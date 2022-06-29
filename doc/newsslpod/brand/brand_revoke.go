package brand

import (
	"context"
	"time"

	"mysslee_qcloud/core/crl"
	"mysslee_qcloud/core/ocsp"
	"mysslee_qcloud/utils"

	log "github.com/sirupsen/logrus"
)

/**
处理缓存中的证书吊销监测
*/

//StartCheckRevoke 检测CA证书吊销情况
func StartCheckRevoke(callback func(string)) {
	ticker := time.NewTicker(12 * time.Hour)
	//开启协程
	go func() {
		defer utils.Recover(nil)
		for range ticker.C {
			go CheckRevoke(callback)
			sendCABrandInfo()
		}
	}()
}

//CheckRevoke 检测吊销 /采用多线程进行查询
func CheckRevoke(callback func(string)) {
	for brand := range ocspChan { //从ocsp管道中获取信息
		//先通过OCSP方式检测 ，如果没有ocsp采用CRL检测
		go func(brand *CABrand) {
			defer utils.Recover(nil)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			defer cancel()
			if brand.Revoke { //如果该证书以吊销，不再进行检测
				return
			}
			if ocspraw := ocsp.GetOCSPdata(ctx, brand.X509, nil); ocspraw != nil {
				rocsp, err := ocsp.ParseResponse(*ocspraw, nil)
				if err == nil {
					if rocsp.Status == ocsp.Revoked { //如果吊销信息为Revoke,那么将CA证书的吊销情况设置为true
						brand.Revoke = true
						// 如果证书被吊销了，回调通知方法
						if callback != nil {
							callback(brand.Hash)
						}
						log.WithFields(log.Fields{
							"hash":     brand.Hash,
							"certName": brand.CertName,
							"line":     utils.ShowCallerMessage(1),
							"event":    "使用ocsp检测证书吊销情况",
						}).Errorf("hash:%v,certName:%v,has Revoked", brand.Hash, brand.CertName)
					}
				}
			} else {
				revoke, err := crl.CheckCRL(ctx, brand.X509)
				if revoke {
					if callback != nil {
						callback(brand.Hash)
					}
				}
				if err != nil {
					log.WithFields(log.Fields{
						"hash":     brand.Hash,
						"certName": brand.CertName,
						"line":     utils.ShowCallerMessage(1),
						"event":    "使用crl检测证书吊销情况",
					}).Warnf("checkCRL 错误:%v", err)
					return
				}
				brand.Revoke = revoke

				if revoke {
					log.WithFields(log.Fields{
						"hash":     brand.Hash,
						"certName": brand.CertName,
						"line":     utils.ShowCallerMessage(1),
						"event":    "证书吊销",
					}).Error("has Revoked")
				}
			}
		}(brand)
	}
}
