package utils

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
	"strings"
)

// SHA1String 生成sha1值(UPPER)
func SHA1String(buf []byte) string {
	s := sha1.New()
	s.Write(buf)
	return strings.ToUpper(hex.EncodeToString(s.Sum(nil)))
}

// SHA256PIN 计算PIN
func SHA256PIN(buf []byte) string {
	s := sha256.Sum256(buf)
	return base64.StdEncoding.EncodeToString(s[:])
}

// SHA1File 计算文件sha1
func SHA1File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	s := sha1.New()
	_, err = io.Copy(s, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(s.Sum(nil)), nil
}
