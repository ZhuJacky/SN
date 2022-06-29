package ssl

func ciphersHaveECC(ciphers []CipherID) bool { //判断是是否有ECC套件
	have := false
	for _, c := range ciphers {
		cipher, exist := c.GetCipherInfo()
		if exist {
			if cipher.Kx == KxECDH {
				have = true
				break
			}
		}
	}
	return have
}

//把ssl2加密套件转换成byte数组
func ssl2CiphersToRawBytes(data []CipherID) []byte {
	var result []byte
	for _, d := range data {
		r := d.ToSSL2CipherRawBytes()
		result = append(result, r...)
	}
	return result
}

//把tls加密套件转换成byte数组
func tlsCiphersToRawBytes(data []CipherID) []byte {
	var result []byte

	for _, d := range data {
		r := d.ToTLSCipherRawBytes()
		result = append(result, r...)
	}
	return result
}
