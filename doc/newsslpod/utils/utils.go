// Package utils provides ...
package utils

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

const (
	CertUpdate  = "0" //不需要证书下载
	CertFromUrl = "1" //需要证书下载
)

var (
	randomStr   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	randomDigit = "0123456789"
)

// RandomCode .
func RandomCode(length int, onlyDigit bool) string {
	buff := &bytes.Buffer{}
	for i := 0; i < length; i++ {
		rand.Seed(rand.Int63() + time.Now().UnixNano())
		if onlyDigit {
			buff.WriteByte(randomDigit[rand.Intn(len(randomDigit))])
		} else {
			buff.WriteByte(randomStr[rand.Intn(len(randomStr))])
		}
	}
	return buff.String()
}

// encrypt password
func EncryptPassword(name, phone, pwd, salt string) string {
	h := sha256.New()
	io.WriteString(h, name)
	io.WriteString(h, salt)
	io.WriteString(h, phone)
	io.WriteString(h, pwd)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// validate password
func VerifyPassword(name, phone, inputPwd, salt, dbPwd string) bool {
	return dbPwd == EncryptPassword(name, phone, inputPwd, salt)
}

// sha1 hash
func SHA1(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return fmt.Sprintf("%X", h.Sum(nil))
}

// sha256 hash
func SHA2(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("%X", h.Sum(nil))
}

// AES-128。key长度：16, 24, 32 bytes 对应 AES-128, AES-192, AES-256
func AESEncrypt(plaintext, key []byte) ([]byte, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := b.BlockSize()
	plaintext = PKCS5Padding(plaintext, bs)
	bm := cipher.NewCBCEncrypter(b, key[:bs])
	ciphertext := make([]byte, len(plaintext))
	bm.CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}

func AESDecrypt(ciphertext, key []byte) ([]byte, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := b.BlockSize()
	bm := cipher.NewCBCDecrypter(b, key[:bs])
	plaintext := make([]byte, len(ciphertext))
	bm.CryptBlocks(plaintext, ciphertext)
	plaintext = PKCS5UnPadding(plaintext)
	return plaintext, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(plaintext []byte) []byte {
	length := len(plaintext)
	unpadding := int(plaintext[length-1])
	return plaintext[:(length - unpadding)]
}

//计算SHA1
func CalculateSHA1(data []byte) string {
	a := sha1.New()
	a.Write(data)
	return fmt.Sprintf("%X", a.Sum(nil))
}

//用于防止go协程中的奔溃，recover
func Recover(ctx context.Context) {
	err := recover()
	if err != nil {
		log.WithFields(log.Fields{
			"message": "异常崩溃" + GetReqInfoFromContext(ctx),
			"stack":   string(debug.Stack()),
		}).Error(err)
	}
}

var random = rand.New(rand.NewSource(time.Now().Unix()))

func RandNonce() int32 {
	return random.Int31()
}

func MaskPhone(pn string) string {
	length := len(pn)
	l := length / 3
	l2 := length % 3
	if l2 > 0 {
		if l2%2 == 0 {
			l = l - l2/2
		} else {
			l = l - (l2+1)/2
		}
	}

	mask := make([]byte, length)
	for i := 0; i < length; i++ {
		if i <= l || i >= length-l {
			mask[i] = pn[i]
		} else {
			mask[i] = '*'
		}
	}
	return string(mask)
}

func MonthToEndDuration(t time.Time) time.Duration {
	t1 := t.Truncate(time.Hour).Add(time.Duration(-t.Hour()) * time.Hour)
	t2 := t1.Add(time.Duration(-int(t1.Day())+1) * 24 * time.Hour)
	t3 := t2.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return t3.Sub(t)
}

func GenerateUUID() string {
	u := uuid.NewV4()
	//剔除uuid中的-
	return strings.Join(strings.Split(u.String(), "-"), "")
}

// int转换成byte数组
func IntToByteArray(data int) ([]byte, error) {
	if data > 65535 {
		return nil, errors.New("输入的数字大于数组的可容纳大小")
	}

	var a []byte = make([]byte, 2)
	binary.BigEndian.PutUint16(a, uint16(data))
	return a, nil
}

//BytesToInt byte到int的转换
func BytesToInt(data []byte) (out int) {

	for _, b := range data {
		out <<= 8
		out += int(b)
	}
	return
}

//将byte数组转换成uint16
func BytesToUint16(data []byte) (out uint16, err error) {
	if len(data) != 2 {
		return 0, errors.New("该方法只支持长度为2的切片")
	}

	out = uint16(data[0])<<8 + uint16(data[1])
	return
}

func Uint16ToByteArray(data uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, data)
	return b
}

//得到调用者信息
func ShowCallerMessage(skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	return fmt.Sprintf("%v:%v", file, line)
}

//Stob 把String转换成bytes
func Stob(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

func CallerStack() string {
	var caller_str string
	for skip := 2; ; skip++ {
		// 获取调用者的信息
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		func_name := runtime.FuncForPC(pc).Name()
		caller_str += "Func : " + func_name + "\nFile:" + file + ":" + fmt.Sprint(line) + "\n"
	}
	return caller_str
}
