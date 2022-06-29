package ct

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	ct "github.com/google/certificate-transparency-go"
	"github.com/google/certificate-transparency-go/asn1"
	"github.com/google/certificate-transparency-go/tls"
	"github.com/google/certificate-transparency-go/x509"
	log "github.com/sirupsen/logrus"
)

const (
	CTUnknown      = 0
	CTQualified    = 1
	CTNotQualified = 2 // CT 不合规
	CTNotAffected  = 3 // 2018/04/30之前的，不受影响
)

func QualifiedDescribe(Qualified int) string {
	switch Qualified {
	case CTUnknown:
		return "未知"
	case CTQualified:
		return "合规"
	case CTNotQualified:
		return "不合规"
	case CTNotAffected:
		return "不受影响"
	}
	return "未知"
}

const (
	CTNotSupport      = ""
	CTFromCertValid   = "来自证书，有效"
	CTFromCertInvalid = "来自证书，无效"
	CTFromTLSValid    = "来自TLS扩展，有效"
	CTFromTLSInvalid  = "来自TLS扩展，无效"
)

type CertificateTransparency struct {
	IsQualified int                           `json:"is_qualified"`
	Description string                        `json:"description"`
	SCTS        []*SignedCertificateTimestamp `json:"scts"`
}

// IsCTQualified 是否合规
func IsCTQualified(sctFromTLS []byte, certs [][]byte) (ctInfo *CertificateTransparency, err error) {
	if len(certs) < 1 {
		err = errors.New("no cert available")
		return
	}

	var leaf, issuer *x509.Certificate
	// 检测证书内的
	leaf, err = x509.ParseCertificate(certs[0])
	if err != nil {
		return
	}

	if len(certs) >= 2 {
		issuer, err = x509.ParseCertificate(certs[1])
		if err != nil {
			return
		}
	}

	ctInfo = &CertificateTransparency{}
	ctInfo.SCTS, err = ParseCertificateTransparencyInfo(sctFromTLS, leaf, issuer)
	if err != nil {
		return
	}

	ctInfo.SCTS = RemoveDuplicatesSCTs(ctInfo.SCTS)

	numRequiredEmbeddedSCTs := NumRequiredEmbeddedSCTs(leaf.NotBefore, leaf.NotAfter)

	// 统计其中有效CT
	var (
		fromTLSOKNotGoogle  int // TLS中不是来自Google的有效SCT
		fromTLSOKByGoogle   int // TLS中来自Google的有效SCT
		fromTLSNotOKCount   int // TLS中不合规的SCT
		fromCertOKNotGoogle int
		fromCertOKByGoogle  int
		fromCertNotOKCount  int
	)
	for _, sct := range ctInfo.SCTS {
		switch sct.Source {
		case CTSourceEmbeddedCert:
			if sct.ValidationStatus == CTValidationVerified {
				if sct.IsOperatedByGoogle {
					fromCertOKByGoogle++
					continue
				}
				fromCertOKNotGoogle++
				continue
			}
			fromCertNotOKCount++
		case CTSourceTLSExtension:
			if sct.ValidationStatus == CTValidationVerified {
				if sct.IsOperatedByGoogle {
					fromTLSOKByGoogle++
					continue
				}
				fromTLSOKNotGoogle++
				continue
			}
			fromTLSNotOKCount++
		}
	}

	// 是否满足内置CT标准
	var isCTFromCertQualified bool
	if fromCertOKByGoogle >= 1 &&
		fromCertOKNotGoogle >= 1 &&
		fromCertOKByGoogle+fromCertOKNotGoogle >= numRequiredEmbeddedSCTs {
		isCTFromCertQualified = true
	}

	// 输出描述
	if isCTFromCertQualified {
		ctInfo.Description = CTFromCertValid
	}

	// 查看是否满足TLS扩展CT合规标准
	var isCTFromTLSQualified bool
	if fromTLSOKByGoogle >= 1 && fromTLSOKNotGoogle >= 1 {
		isCTFromTLSQualified = true
	}

	// 描述
	if (isCTFromCertQualified && fromTLSOKByGoogle+fromTLSOKNotGoogle > 0) || isCTFromTLSQualified {
		if ctInfo.Description != "" {
			ctInfo.Description = ctInfo.Description + "; " + CTFromTLSValid
		} else {
			ctInfo.Description = CTFromTLSValid
		}
	}

	// 判断是否合规

	// 20180430之前无任何CT信息将不受影响，但如果有，但不合规，也视为不合规
	// 如果TLS CT不合规，且又包含了内置SCT，则判断内置CT是否合规
	if leaf.NotBefore.Before(time.Date(2018, 4, 30, 0, 0, 0, 0, time.UTC)) {
		if !isCTFromTLSQualified && !isCTFromCertQualified && fromCertNotOKCount > 0 {
			ctInfo.IsQualified = CTNotQualified
			return
		}
		ctInfo.IsQualified = CTNotAffected
		return
	}

	// 只要有一个合规即为合规
	if isCTFromTLSQualified || isCTFromCertQualified {
		ctInfo.IsQualified = CTQualified
		return
	}

	ctInfo.IsQualified = CTNotQualified
	return
}

type SignedCertificateTimestamp struct {
	LogName            string `json:"log_name"`
	LogID              string `json:"log_id"`
	ValidationStatus   string `json:"validation_status"`
	Source             string `json:"source"`
	IssuedAt           string `json:"issued_at"`
	HashAlgo           string `json:"hash_algo"`
	SignAlgo           string `json:"sign_algo"`
	SignatureData      string `json:"signature_data"`
	IsOperatedByGoogle bool   `json:"is_operated_by_google"`
}

func (s *SignedCertificateTimestamp) String() string {
	return fmt.Sprintf("\nLogName: %v\nLogID: %v\nValidationStatus: %v\nSource: %v\nIssuedAt: %v\nHashAlgo: %v\nSignAlgo: %v\nSignatureData: %v\n",
		s.LogName, s.LogID, s.ValidationStatus, s.Source, s.IssuedAt, s.HashAlgo, s.SignAlgo, s.SignatureData)
}

const (
	CTSourceEmbeddedCert = "Embedded in certificate" // sct 内置证书
	CTSourceTLSExtension = "TLS extension"           // sct 来自tls扩展
)

func ParseCertificateTransparencyInfo(sctFromTLS []byte, leaf, issuer *x509.Certificate) (info []*SignedCertificateTimestamp, err error) {
	if leaf == nil {
		err = errors.New("Parse CT info : no cert available")
		return
	}

	info = []*SignedCertificateTimestamp{}

	// 验证内嵌的SCT需要颁发者信息
	if issuer != nil {
		for _, ext := range leaf.Extensions {
			if ext.Id.Equal(asn1.ObjectIdentifier(X509SCTExtensionID)) {
				scts, err := ParseX509SCTExtension(ext.Value)
				if err != nil {
					log.Warnf(`Parse sct from x509 err: %v %v`, err, hex.Dump(ext.Value))
					err = nil
					break
				}
				for _, sct := range scts {
					sctParse, err := verify(sct, leaf, issuer, CTSourceEmbeddedCert)
					if err == nil {
						info = append(info, sctParse)
						continue
					}

					log.Warnf("Verify sct err %v", err)
				}
				break
			}
		}
	}

	if len(sctFromTLS) > 0 {
		scts, err := ParseTLSSCTExtension(sctFromTLS)
		if err != nil {
			log.Warnf("Parse sct from tls err: %v %v", err, hex.Dump(sctFromTLS))
			return info, nil
		}
		for _, sct := range scts {
			sctParse, err := verify(sct, leaf, issuer, CTSourceTLSExtension)
			if err == nil {
				info = append(info, sctParse)
				continue
			}

			log.Warnf("Verify sct err %v", err)
		}
	}

	return
}

const (
	PublicKeyPrefix = "-----BEGIN PUBLIC KEY-----\r\n"
	PublicKeySuffix = "\r\n-----END PUBLIC KEY-----"
)

const (
	CTValidationVerified        = "Verified"                // sct 信息验证有效
	CTValidatioaLogDisqualified = "Log Server Disqualified" // ct 日志服务不合规
	CTValidationVerifiedFail    = "Not Qualified"           // sct 验证错误
	CTValidationFromUnknowLog   = "From unknown log"        // 未知的CT服务商
)

func verify(sct ct.SignedCertificateTimestamp, leaf, issuer *x509.Certificate, from string) (out *SignedCertificateTimestamp, err error) {
	out = &SignedCertificateTimestamp{}

	// 从CT日志表中查询日志服务商
	out.LogID = strings.ToUpper(hex.EncodeToString(sct.LogID.KeyID[:]))
	out.IssuedAt = time.Unix(int64(sct.Timestamp/1000), int64(sct.Timestamp%1000)).Local().Format("2006-01-02 15:04:05 MST")
	out.HashAlgo = sct.Signature.Algorithm.Hash.String()
	out.SignAlgo = sct.Signature.Algorithm.Signature.String()
	out.SignatureData = strings.ToUpper(hex.EncodeToString(sct.Signature.Signature))
	out.Source = from

	logIDStr := base64.StdEncoding.EncodeToString(sct.LogID.KeyID[:])
	ctLogInfo, exist := CTLogList[logIDStr]
	if !exist {
		out.ValidationStatus = CTValidationFromUnknowLog
		return
	}
	out.LogName = ctLogInfo.Description

	for _, optID := range ctLogInfo.OperatedBy {
		if operator, ok := CTOperators[optID]; ok && operator == "Google" {
			out.IsOperatedByGoogle = true
			break
		}
	}

	pubKeyPEM := PublicKeyPrefix + ctLogInfo.Key + PublicKeySuffix
	pubkey, _ /* keyhash */, rest, err := ct.PublicKeyFromPEM([]byte(pubKeyPEM))
	if err != nil {
		return
	}
	if len(rest) > 0 {
		err = errors.New("extra data found after PEM key decoded")
		return
	}

	merkleLeaf := &ct.MerkleTreeLeaf{}
	switch from {
	case CTSourceTLSExtension:
		merkleLeaf, err = ct.MerkleTreeLeafFromChain([]*x509.Certificate{leaf}, ct.X509LogEntryType, sct.Timestamp)
		if err != nil {
			return
		}
	case CTSourceEmbeddedCert:
		if issuer == nil {
			err = errors.New("verify sct embedded in cert need issuer cert")
			return
		}
		merkleLeaf, err = ct.MerkleTreeLeafForEmbeddedSCT([]*x509.Certificate{leaf, issuer}, sct.Timestamp)
		if err != nil {
			return
		}
	default:
		return nil, errors.New("unknown sct source")
	}

	logEntry := ct.LogEntry{Leaf: *merkleLeaf}
	sctData, err := ct.SerializeSCTSignatureInput(sct, logEntry)
	if err != nil {
		return
	}

	out.ValidationStatus = CTValidationVerified
	if err := tls.VerifySignature(pubkey, sctData, tls.DigitallySigned(sct.Signature)); err != nil {
		out.ValidationStatus = CTValidationVerifiedFail
	}

	// 如果证书签发时间超过了ct被标记为无效的时候之后，该ct视为无效
	if ctLogInfo.DisqualifiedAt != 0 {
		invalidTime := time.Unix(ctLogInfo.DisqualifiedAt, 0)
		if leaf.NotBefore.UTC().Sub(invalidTime) >= 0 {
			out.ValidationStatus = CTValidatioaLogDisqualified
			return
		}
	}
	return
}

// RemoveDuplicatesSCTs 去重
// Sort the embedded log IDs and remove duplicates, so that only a single
// SCT from each log is accepted. This is to handle the case where a given
// log returns different SCTs for the same precertificate (which is
// permitted, but advised against)
func RemoveDuplicatesSCTs(scts []*SignedCertificateTimestamp) (out []*SignedCertificateTimestamp) {

	out = []*SignedCertificateTimestamp{}

	for i := 0; i < len(scts); i++ {

		sct := scts[i]

		var hasExist bool
		for _, v := range out {
			if v.Source == sct.Source && v.LogID == sct.LogID {
				hasExist = true
				break
			}
		}

		if !hasExist {
			out = append(out, sct)
		}
	}
	return
}

// NumRequiredEmbeddedSCTs 需要的CT数量
func NumRequiredEmbeddedSCTs(begin, end time.Time) int {
	lifeTimeInMonths, hasPartialMonth := RoundedDownMonthDifference(begin, end)

	num := 5
	if lifeTimeInMonths > 39 || (lifeTimeInMonths == 39 && hasPartialMonth) {
		num = 5
	} else if lifeTimeInMonths > 27 || (lifeTimeInMonths == 27 && hasPartialMonth) {
		num = 4
	} else if lifeTimeInMonths >= 15 {
		num = 3
	} else {
		num = 2
	}

	return num
}

// RoundedDownMonthDifference 月差
// 参考 https://chromium.googlesource.com/chromium/src/+/lkgr/components/certificate_transparency/chrome_ct_policy_enforcer.cc
func RoundedDownMonthDifference(begin, end time.Time) (monthDiff int, hasPartialMonth bool) {
	if begin.After(end) {
		monthDiff = 0
		hasPartialMonth = false
		return
	}

	hasPartialMonth = true
	monthDiff = (end.Year()-begin.Year())*12 + int(end.Month()-begin.Month())

	if monthDay(end) < monthDay(begin) {
		monthDiff--
	} else if monthDay(end) == monthDay(begin) {
		hasPartialMonth = false
	}

	return
}

func monthDay(t time.Time) int {
	_, _, d := t.Date()
	return d
}
