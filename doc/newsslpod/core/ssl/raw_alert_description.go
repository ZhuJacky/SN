package ssl

const (
	CloseNotify               = 0
	UnexpectedMessage         = 10
	BadRecordMac              = 20
	DecryptionFailRESERVED    = 21
	RecordOverflow            = 22
	DecompressionFailure      = 30
	HandshakeFailure          = 40
	NoCertificateRESERVED     = 41
	BadCertificate            = 42
	UnsupportedCertificate    = 43
	CertificateRevoked        = 44
	CertificateExpired        = 45
	CertificateUnknown        = 46
	IllegalParamerter         = 47
	UnknownCa                 = 48
	AccessDenied              = 49
	DecodeError               = 50
	DecryptError              = 51
	ExportRestrictionRESERVED = 60
	ProtocolVersion           = 70
	InsufficientSecurity      = 71
	InternalError             = 80
	InappropriateFallback     = 86
	UserCanceled              = 90
	NoRenegotiation           = 100
	UnsupportedExtension      = 110
	UnrecognizedName          = 112
	NoApplicationProtocol     = 120
	AlertSuccess              = 255
)

const (
	Warning byte = 1
	Fatal   byte = 2
)
