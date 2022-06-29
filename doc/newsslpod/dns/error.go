// Package dns provides ...
package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/miekg/dns"
)

var (
	ErrPingDNSServer = errors.New("ping server in few seconds timeout")
	ErrEmptyEDNSAddr = errors.New("need specify edns0 addr")
	ErrTooManyQuery  = errors.New("Lookup reaches the maximum number of " + strconv.Itoa(MaxParse))
	ErrTooManyCNAME  = errors.New("CNAME reaches the maximum number of " + strconv.Itoa(MaxCNAME))
)

const detailDNSTimeout = "时间超时"
const detailDNSNetFailure = "网络错误"
const detailServerFailure = "数据错误"
const detailNameError = "域名错误"

// DNSError wraps a DNS error with various relevant information
type DNSError struct {
	recordType uint16
	hostname   string
	// Exactly one of rCode or underlying should be set.
	underlying error
	rCode      int
}

func (d DNSError) Error() string {
	var detail string
	if d.underlying != nil {
		if netErr, ok := d.underlying.(*net.OpError); ok {
			if netErr.Timeout() {
				detail = detailDNSTimeout
			} else {
				detail = detailDNSNetFailure
			}
			// Note: we check d.underlying here even though `Timeout()` does this because the call to `netErr.Timeout()` above only
			// happens for `*net.OpError` underlying types!
		} else if d.underlying == context.Canceled || d.underlying == context.DeadlineExceeded {
			detail = detailDNSTimeout
		} else {
			detail = detailServerFailure
		}
	} else if d.rCode != dns.RcodeSuccess {
		if d.rCode == dns.RcodeNameError {
			detail = detailNameError
		} else {
			detail = dns.RcodeToString[d.rCode]
		}
	} else {
		detail = detailServerFailure
	}
	// return fmt.Sprintf("DNS problem: %s looking up %s for %s", detail,
	// dns.TypeToString[d.recordType], d.hostname)
	return fmt.Sprintf("DNS 错误：查询 %s %s", dns.TypeToString[d.recordType], detail)
}

// Timeout returns true if the underlying error was a timeout
func (d DNSError) Timeout() bool {
	if netErr, ok := d.underlying.(*net.OpError); ok {
		return netErr.Timeout()
	} else if d.underlying == context.Canceled || d.underlying == context.DeadlineExceeded {
		return true
	}
	return false
}
