package ssl

import "fmt"

type CodeSignature byte

type CodeHash byte

type SignatureHashAlgorithm struct {
	Hash      CodeHash
	Signature CodeSignature
}

func (s SignatureHashAlgorithm) String() string {
	return string(fmt.Sprintf("%v/%v", s.Hash.String(), s.Signature.String()))
}

const (
	Signature_NULL     CodeSignature = 0x00
	Signature_RSA      CodeSignature = 0x01
	Signature_DSA      CodeSignature = 0x02
	Signature_ECDSA    CodeSignature = 0x03
	Signature_Unknown4 CodeSignature = 0x04
	Signature_Unknown5 CodeSignature = 0x05
	Signature_Unknown6 CodeSignature = 0x06
	Signature_ED25519  CodeSignature = 0x07
	Signature_ED448    CodeSignature = 0x08
	Signature_Unknown9 CodeSignature = 0x09
	Signature_Unknowna CodeSignature = 0x0a
	Signature_Unknownb CodeSignature = 0x0b
)

func (c CodeSignature) String() string {
	switch c {
	case Signature_NULL:
		return "NULL"
	case Signature_RSA:
		return "RSA"
	case Signature_DSA:
		return "DSA"
	case Signature_ECDSA:
		return "ECDSA"
	case Signature_Unknown4:
		return "UNKNOWN4"
	case Signature_Unknown5:
		return "UNKNOWN5"
	case Signature_Unknown6:
		return "UNKNOWN6"
	case Signature_ED25519:
		return "ED25519"
	case Signature_ED448:
		return "ED448"
	default:
		return "UNKNOWN"
	}
}

const (
	Hash_NULL      CodeHash = 0x00
	Hash_MD5       CodeHash = 0x01
	Hash_SHA1      CodeHash = 0x02
	Hash_SHA224    CodeHash = 0x03
	Hash_SHA256    CodeHash = 0x04
	Hash_SHA384    CodeHash = 0x05
	Hash_SHA512    CodeHash = 0x06
	Hash_Intrinsic CodeHash = 0x08
)

func (c CodeHash) String() string {
	switch c {
	case Hash_NULL:
		return "NULL"
	case Hash_MD5:
		return "MD5"
	case Hash_SHA1:
		return "SHA1"
	case Hash_SHA224:
		return "SHA224"
	case Hash_SHA256:
		return "SHA256"
	case Hash_SHA384:
		return "SHA384"
	case Hash_SHA512:
		return "SHA512"
	case Hash_Intrinsic:
		return "INTRINSIC"
	default:
		return "UNKNOWN"
	}
}

var defaultSignatureHashAlgorithms []byte
var allSignatureHashAlgorithms []byte

func GenerateSignatureHashAlgorithms(algorithms []SignatureHashAlgorithm) []byte {
	data := make([]byte, 0)
	for _, a := range algorithms {
		data = append(data, byte(a.Hash))
		data = append(data, byte(a.Signature))
	}
	return data
}

func GetDefaultSignatureHashAlgorithms() []byte {
	if len(defaultSignatureHashAlgorithms) > 0 {
		return defaultSignatureHashAlgorithms
	}

	data := make([]SignatureHashAlgorithm, 0)

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA512,
		Signature: Signature_RSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA512,
		Signature: Signature_DSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA512,
		Signature: Signature_ECDSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA384,
		Signature: Signature_RSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA384,
		Signature: Signature_DSA,
	})

	data = append(data, SignatureHashAlgorithm{

		Hash:      Hash_SHA384,
		Signature: Signature_ECDSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA256,
		Signature: Signature_RSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA256,
		Signature: Signature_DSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA256,
		Signature: Signature_ECDSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA224,
		Signature: Signature_RSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA224,
		Signature: Signature_DSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA224,
		Signature: Signature_ECDSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA1,
		Signature: Signature_RSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA1,
		Signature: Signature_DSA,
	})
	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA1,
		Signature: Signature_ECDSA,
	})

	defaultSignatureHashAlgorithms = GenerateSignatureHashAlgorithms(data)

	return defaultSignatureHashAlgorithms
}

func GetAllSignatureHashAlgorithms() []byte {
	if len(allSignatureHashAlgorithms) > 0 {
		return allSignatureHashAlgorithms
	}

	data := make([]SignatureHashAlgorithm, 0)

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_Intrinsic,
		Signature: Signature_Unknown4,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_Intrinsic,
		Signature: Signature_ED448,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA512,
		Signature: Signature_RSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA512,
		Signature: Signature_DSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA512,
		Signature: Signature_ECDSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA384,
		Signature: Signature_RSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA384,
		Signature: Signature_DSA,
	})

	data = append(data, SignatureHashAlgorithm{

		Hash:      Hash_SHA384,
		Signature: Signature_ECDSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA256,
		Signature: Signature_RSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA256,
		Signature: Signature_DSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA256,
		Signature: Signature_ECDSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA224,
		Signature: Signature_RSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA224,
		Signature: Signature_DSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA224,
		Signature: Signature_ECDSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA1,
		Signature: Signature_RSA,
	})

	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA1,
		Signature: Signature_DSA,
	})
	data = append(data, SignatureHashAlgorithm{
		Hash:      Hash_SHA1,
		Signature: Signature_ECDSA,
	})

	allSignatureHashAlgorithms = GenerateSignatureHashAlgorithms(data)
	return allSignatureHashAlgorithms
}
