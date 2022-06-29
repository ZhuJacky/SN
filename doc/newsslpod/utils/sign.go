// Package sign provides ...
package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// 签名验证机制
// 排序参数 -> hmacsha256 -> base64
func SignValues(vals url.Values, secretKey string) (string, string) {
	// 手动排序
	keys := make([]string, 0, len(vals))
	for k := range vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	for _, k := range keys {
		vs := vals[k]
		prefix := k + "="
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(prefix)
			buf.WriteString(v)
		}
	}
	mac := hmac.New(sha256.New, []byte(secretKey))
	return fmt.Sprintf("%x", mac.Sum(buf.Bytes())), buf.String()
}

// reportapp sign method
// 将基本参数拼接为特定字符串 s1 -> hmac-sha1 签名
// partnerId=xxx&timestamp=xxx&f=1&count=0&expire=xxx&domain=xxx&port=xxx&ip=xxx
type KV struct {
	Key   string
	Value interface{}
}

func SignReport(secretKey string, params []KV) (string, string) {
	tmp := make([]string, len(params))
	for i, kv := range params {
		tmp[i] = fmt.Sprintf("%s=%v", kv.Key, kv.Value)
	}
	plaintext := strings.Join(tmp, "&")

	h := hmac.New(sha1.New, []byte(secretKey))
	h.Write([]byte(plaintext))
	return plaintext, hex.EncodeToString(h.Sum(nil))
}
