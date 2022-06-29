// Package utils provides ...
package utils

import (
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/weppos/publicsuffix-go/publicsuffix"
)

// CheckPort 验证端口
func CheckPort(port string) bool {
	p, err := strconv.Atoi(port)
	return err == nil && (p > 0 && p <= 65535)
}

var regexCheckIP = regexp.MustCompile(
	`^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$`)

// ValidateIP 验证IP
func ValidateIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

var regPhone = regexp.MustCompile(`^(13[0-9]|14[579]|15[0-3,5-9]|16[6]|17[0135678]|18[0-9]|19[89])\d{8}$`)

// ValidatePhone ValidatePhone
func ValidatePhone(p string) bool {
	return regPhone.MatchString(p)
}

// check email
var regEmail = regexp.MustCompile(`^[\w._%+-]+@[\w.-]+\.[a-zA-Z]{2,4}$`)

// ValidateEmail ValidateEmail
func ValidateEmail(e string) bool {
	return regEmail.MatchString(e)
}

// var regexDomain = regexp.MustCompile(`^(\*\.)?([A-Za-z0-9_\-一-龥]{1,63}\.)*([A-Za-z0-9_\-一-龥]{1,256}\.[A-Za-z一-龥]{1,256})$`)
// 原始 `^[0-9\p{L}][0-9\p{L}-\.]{1,61}[0-9\p{L}]\.[0-9\p{L}][\p{L}-]*[0-9\p{L}]+$`
// 开放下划线 _
var regDomain = regexp.MustCompile(`^[0-9\p{L}_][0-9\p{L}-\._]{0,250}\.[0-9\p{L}][\p{L}-]*[0-9\p{L}]+$`)

// ValidateDomain ValidateDomain
func ValidateDomain(d string) bool {
	return regDomain.MatchString(d)
}

var regWildcardDomain = regexp.MustCompile(`^[0-9\p{L}_\*][0-9\p{L}-\._]{0,250}\.[0-9\p{L}][\p{L}-]*[0-9\p{L}]+$`)

// ValidateWildcardDomain ValidateWildcardDomain
func ValidateWildcardDomain(d string) bool {
	return regWildcardDomain.MatchString(d)
}

// 验证密码合法性
// 1、不能包含空格
// 2、字母、数字、字符至少包含两种
// 3、长度在 6-16
var (
	regNumber    = regexp.MustCompile(`\d`)
	regLetter    = regexp.MustCompile(`(?i)[a-z]`)
	regCharacter = regexp.MustCompile(`[^\da-zA-Z\s]`)

	regTmpPassword = regexp.MustCompile(`^[\w~!@#$%^&*()_+-=<>,./?'";:\[\]\{\}\|\\]{6,24}$`)
)

// ValidatePassword ValidatePassword
func ValidatePassword(pwd string) bool {
	// if len(pwd) < 6 || len(pwd) > 24 {
	// 	return false
	// }
	// if strings.Contains(pwd, " ") {
	// 	return false
	// }
	// return regNumber.MatchString(pwd) && regLetter.MatchString(pwd) ||
	// 	regNumber.MatchString(pwd) && regCharacter.MatchString(pwd) ||
	// 	regLetter.MatchString(pwd) && regCharacter.MatchString(pwd)

	return regTmpPassword.MatchString(pwd)
}

// tag
var regTagName = regexp.MustCompile(`^[\w\p{Han}]{1,10}$`)

// ValidateTagName ValidateTagName
func ValidateTagName(tn string) bool {
	return regTagName.MatchString(tn)
}

// csp token
var regCSPToken = regexp.MustCompile(`^\w{32}$`)

// ValidateCSPToken ValidateCSPToken
func ValidateCSPToken(t string) bool {
	return regCSPToken.MatchString(t)
}

// ValidatePort 验证端口
func ValidatePort(port string) bool {
	p, err := strconv.Atoi(port)
	return err == nil && p >= 0 && p <= 65535
}

// 验证u2f命名
var regU2FName = regexp.MustCompile(`^[\w\p{Han}]{1,10}$`)

// ValidateU2FName ValidateU2FName
func ValidateU2FName(name string) bool {
	return regU2FName.MatchString(name)
}

var regDomain2 = regexp.MustCompile(`^[\w-_]{1,63}$`)

// ValidateDomain2 新版验证域名
func ValidateDomain2(d string) bool {
	// 最大长度为 255
	if len(d) > 255 || len(d) < 3 {
		return false
	}
	// test.hello.example.com
	// DomainName{"com", "example", "test.hello"}
	// TLD SLD TRD
	dn, err := publicsuffix.ParseFromListWithOptions(publicsuffix.DefaultList, d, &publicsuffix.FindOptions{
		IgnorePrivate: false,
		DefaultRule:   &publicsuffix.Rule{},
	})
	if err != nil {
		return false
	}
	if !regDomain2.MatchString(dn.SLD) ||
		strings.HasPrefix(dn.SLD, "-") ||
		strings.HasSuffix(dn.SLD, "-") {
		return false
	}
	if dn.TRD != "" {
		for _, v := range strings.Split(dn.TRD, ".") {
			if !regDomain2.MatchString(v) ||
				strings.HasPrefix(v, "-") ||
				strings.HasSuffix(v, "-") {
				return false
			}
		}
	}
	return true
}
