package ssl

import (
	"strconv"

	log "github.com/sirupsen/logrus"
)

func CiphersFromHex(data string) []CipherID {
	var cipherHexs []string
	var ciphers []CipherID
	var unSupportCiphers []string
	for i := 0; i <= len(data)-4; i += 4 {
		cipher := data[i : i+4]
		if len(cipher) < 4 {
			continue
		}
		cipherHexs = append(cipherHexs, cipher)
	}

	var find bool
	for _, cip := range cipherHexs {
		for k, c := range cipherInfos {
			data, err := strconv.ParseUint(cip, 16, 64)
			if err != nil {
				unSupportCiphers = append(unSupportCiphers, cip)
				break
			}
			if c.Data == CipherID(data) {
				ciphers = append(ciphers, k)
				find = true
				break
			}
		}

		if !find {
			unSupportCiphers = append(unSupportCiphers, cip)

		}

		find = false
	}

	if len(unSupportCiphers) > 0 {
		panic("FromHex: 有不支持的加密套件！")
	}
	return ciphers
}

func StringToCiphers(data string) ([]string, []string) {
	var cipherHexs []string
	var ciphers []string
	var unSupportCiphers []string
	for i := 0; i <= len(data)-4; i += 4 {
		cipher := data[i : i+4]
		if len(cipher) < 4 {
			continue
		}
		cipherHexs = append(cipherHexs, cipher)
	}

	var find bool
	for _, cip := range cipherHexs {
		for _, c := range cipherInfos {

			data, err := strconv.ParseUint(cip, 16, 16)
			if err != nil {
				unSupportCiphers = append(unSupportCiphers, cip)
				continue
			}

			if c.Data == CipherID(data) {
				ciphers = append(ciphers, c.Name)
				find = true
				break
			}
		}

		if !find {
			log.Print("StringToCiphers ", cip)
			unSupportCiphers = append(unSupportCiphers, cip)
		}
		find = false
	}
	return ciphers, unSupportCiphers
}

// CiphersToRawBytes 转化加密套件到Raw数组
func CiphersToRawBytes(ciphers []uint16) []byte {
	var b = make([]byte, len(ciphers)*2)
	for i := 0; i < len(ciphers); i++ {
		b[i*2] = byte(ciphers[i] >> 8)
		b[i*2+1] = byte(ciphers[i])
	}
	return b
}

// 致命的SSL警告
func IsFatalAlertMsg(msg []byte) bool {
	return len(msg) == 2 && msg[0] == 0x02 /*Fatal*/
}
