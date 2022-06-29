package ssl

func AlertToString(alertCode int) string {
	switch alertCode {
	case CloseNotify:
		return "close_notify"
	case UnexpectedMessage:
		return "unexpected_message"
	case BadRecordMac:
		return "bad_record_mac"
	case DecryptionFailRESERVED:
		return "decryption_fail_reserved"
	case RecordOverflow:
		return "record_overflow"
	case DecompressionFailure:
		return "decompression_failure"
	case HandshakeFailure:
		return "handshake_failure"
	case NoCertificateRESERVED:
		return "no_certificate_reserved"
	case BadCertificate:
		return "bad_certificate"
	case UnsupportedCertificate:
		return "unsupported_certificate"
	case CertificateRevoked:
		return "certificate_revoked"
	case CertificateExpired:
		return "certificate_expired"
	case CertificateUnknown:
		return "certificate_unknown"
	case IllegalParamerter:
		return "illegal_paramerter"
	case UnknownCa:
		return "unknown_ca"
	case AccessDenied:
		return "access_denied"
	case DecodeError:
		return "decode_error"
	case DecryptError:
		return "decrypt_error"
	case ExportRestrictionRESERVED:
		return "export_restriction_reserved"
	case ProtocolVersion:
		return "protocol_version"
	case InsufficientSecurity:
		return "insufficient_security"
	case InternalError:
		return "internal_error"
	case InappropriateFallback:
		return "inappropriate_fallback"
	case UserCanceled:
		return "user_canceled"
	case NoRenegotiation:
		return "no_renegotiation"
	case UnsupportedExtension:
		return "unsupported_extension"
	case UnrecognizedName:
		return "unreognized_name"
	case NoApplicationProtocol:
		return "no_application_protocol"
	case AlertSuccess:
		return "alert_success"
	default:
		return "unknown"
	}
}
