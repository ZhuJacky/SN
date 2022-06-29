// Package dns provides ...
package dns

import (
	"context"
	"crypto/tls"
	"fmt"
	"mysslee_qcloud/utils"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/weppos/publicsuffix-go/publicsuffix"
)

// DnsAAAAContextKey TODO
type DnsAAAAContextKey struct{}

// DnsServerContextKey TODO
type DnsServerContextKey struct{}

// const
const (
	DOWN = "down"
	UP   = "up"
)

func parseCidr(network string, comment string) net.IPNet {
	_, net, err := net.ParseCIDR(network)
	if err != nil {
		panic(fmt.Sprintf("error parsing %s (%s): %s", network, comment, err))
	}
	return *net
}

var (
	// Private CIDRs to ignore
	privateNetworks = []net.IPNet{
		// tencent private network
		{
			IP:   []byte{9, 0, 0, 0},
			Mask: []byte{255, 0, 0, 0},
		},
		{
			IP:   []byte{11, 0, 0, 0},
			Mask: []byte{255, 0, 0, 0},
		},
		{
			IP:   []byte{30, 0, 0, 0},
			Mask: []byte{255, 0, 0, 0},
		},
		// RFC1918
		// 10.0.0.0/8
		{
			IP:   []byte{10, 0, 0, 0},
			Mask: []byte{255, 0, 0, 0},
		},
		// 172.16.0.0/12
		{
			IP:   []byte{172, 16, 0, 0},
			Mask: []byte{255, 240, 0, 0},
		},
		// 192.168.0.0/16
		{
			IP:   []byte{192, 168, 0, 0},
			Mask: []byte{255, 255, 0, 0},
		},
		// RFC5735
		// 127.0.0.0/8
		{
			IP:   []byte{127, 0, 0, 0},
			Mask: []byte{255, 0, 0, 0},
		},
		// RFC1122 Section 3.2.1.3
		// 0.0.0.0/8
		{
			IP:   []byte{0, 0, 0, 0},
			Mask: []byte{255, 0, 0, 0},
		},
		// RFC3927
		// 169.254.0.0/16
		{
			IP:   []byte{169, 254, 0, 0},
			Mask: []byte{255, 255, 0, 0},
		},
		// RFC 5736
		// 192.0.0.0/24
		{
			IP:   []byte{192, 0, 0, 0},
			Mask: []byte{255, 255, 255, 0},
		},
		// RFC 5737
		// 192.0.2.0/24
		{
			IP:   []byte{192, 0, 2, 0},
			Mask: []byte{255, 255, 255, 0},
		},
		// 198.51.100.0/24
		{
			IP:   []byte{192, 51, 100, 0},
			Mask: []byte{255, 255, 255, 0},
		},
		// 203.0.113.0/24
		{
			IP:   []byte{203, 0, 113, 0},
			Mask: []byte{255, 255, 255, 0},
		},
		// RFC 3068
		// 192.88.99.0/24
		{
			IP:   []byte{192, 88, 99, 0},
			Mask: []byte{255, 255, 255, 0},
		},
		// RFC 2544
		// 192.18.0.0/15
		{
			IP:   []byte{192, 18, 0, 0},
			Mask: []byte{255, 254, 0, 0},
		},
		// RFC 3171
		// 224.0.0.0/4
		{
			IP:   []byte{224, 0, 0, 0},
			Mask: []byte{240, 0, 0, 0},
		},
		// RFC 1112
		// 240.0.0.0/4
		{
			IP:   []byte{240, 0, 0, 0},
			Mask: []byte{240, 0, 0, 0},
		},
		// RFC 919 Section 7
		// 255.255.255.255/32
		{
			IP:   []byte{255, 255, 255, 255},
			Mask: []byte{255, 255, 255, 255},
		},
		// RFC 6598
		// 100.64.0.0./10
		{
			IP:   []byte{100, 64, 0, 0},
			Mask: []byte{255, 192, 0, 0},
		},
	}
	// Sourced from https://www.iana.org/assignments/iana-ipv6-special-registry/iana-ipv6-special-registry.xhtml
	// where Global, Source, or Destination is False
	privateV6Networks = []net.IPNet{
		parseCidr("::/128", "RFC 4291: Unspecified Address"),
		parseCidr("::1/128", "RFC 4291: Loopback Address"),
		parseCidr("::ffff:0:0/96", "RFC 4291: IPv4-mapped Address"),
		parseCidr("100::/64", "RFC 6666: Discard Address Block"),
		parseCidr("2001::/23", "RFC 2928: IETF Protocol Assignments"),
		parseCidr("2001:2::/48", "RFC 5180: Benchmarking"),
		parseCidr("2001:db8::/32", "RFC 3849: Documentation"),
		parseCidr("2001::/32", "RFC 4380: TEREDO"),
		parseCidr("fc00::/7", "RFC 4193: Unique-Local"),
		parseCidr("fe80::/10", "RFC 4291: Section 2.5.6 Link-Scoped Unicast"),
		parseCidr("ff00::/8", "RFC 4291: Section 2.7"),
		// We disable validations to IPs under the 6to4 anycase prefix because
		// there's too much risk of a malicious actor advertising the prefix and
		// answering validations for a 6to4 host they do not control.
		// https://community.letsencrypt.org/t/problems-validating-ipv6-against-host-running-6to4/18312/9
		parseCidr("2002::/16", "RFC 7526: 6to4 anycast prefix deprecated"),
	}
)

// IsPrivateIP if s is not valid IPv4 or IPv6, return false
// eg. 192.168.99.100
//     2001:db8::68
func IsPrivateIP(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			return IsPrivateV4(ip)
		}
	}
	return IsPrivateV6(ip)
}

// IsPrivateV4 determine if ip is IPv4
func IsPrivateV4(ip net.IP) bool {
	for _, net := range privateNetworks {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}

// IsPrivateV6 determine if ip is IPv6
func IsPrivateV6(ip net.IP) bool {
	for _, net := range privateV6Networks {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}

// DNSServer dns server stat
type DNSServer struct {
	Addr         string // the complete addr, contain port
	LocationCode string // eg 86
	SupportEDNS  bool   // support edns
	State        string // up, stop, unknown
}

// complete server addr, if don't have port will fill 53.
func completeServers(servers []string) (dss []*DNSServer) {
	if len(servers) == 0 {
		return nil
	}
	dss = make([]*DNSServer, len(servers))
	for i, v := range servers {
		if index := strings.Index(v, ":"); index == -1 {
			v += ":53"
		}
		dss[i] = &DNSServer{Addr: v, State: UP}
	}
	return
}

// DNSClient the client
type DNSClient struct {
	lock                     sync.Mutex                         // lock
	client                   dns.Client                         // dns client
	servers                  []*DNSServer                       // DNS server address
	allowRestrictedAddresses bool                               // 是否允许"限制地址"
	maxTries                 int                                // 最大尝试次数
	queryAAAA                bool                               // 是否查询 ipv6 地址
	callback                 func(err error, value interface{}) // callback function
}

// default client optional
var (
	config, _       = dns.ClientConfigFromFile("/etc/resolv.conf")
	defaultCallback = func(err error, value interface{}) {
		log.Println("dns callback", err, value)
	}
	defaultOpt = &ClientOpt{
		Net:         "udp",
		ReadTimeout: time.Second * 10,
		MaxTries:    1,
		QueryAAAA:   false,
		Servers:     config.Servers,
		Callback:    defaultCallback,
	}
)

// ClientOpt the option to new client
type ClientOpt struct {
	Net         string        // tcp, tcp-tls, udp(default)
	ReadTimeout time.Duration // how long
	MaxTries    int           // the max query count
	QueryAAAA   bool          // maybe you don't want to query IPv6
	Servers     []string      // dns server addr
	Callback    func(err error, value interface{})
}

// DefaultClient default client
var DefaultClient = NewDNSClient(defaultOpt, nil)

// ExtraOpt extra optional
type ExtraOpt struct {
	Server    string // specify the dns addr
	EDNS      bool   // use edns
	LocalAddr string // the edns lcality ip
	AAAA      bool   // query ipv6
}

// NewDNSClient new dns client
func NewDNSClient(opt *ClientOpt, tlsConfig *tls.Config) *DNSClient {
	client := &DNSClient{
		client: dns.Client{
			Net:         opt.Net,
			ReadTimeout: opt.ReadTimeout,
			TLSConfig:   tlsConfig,
			// SingleInflight: false,
		},
		maxTries:                 opt.MaxTries,
		queryAAAA:                opt.QueryAAAA,
		allowRestrictedAddresses: false,
		servers:                  completeServers(opt.Servers),
		callback:                 opt.Callback,
	}
	// prom初始化dns服务器数量
	PromDNSOnline.Set(float64(len(opt.Servers)))
	go client.Ping()
	return client
}

// SetServers change server addr
// <locationcode>, <service addr>, <support edns>
func (cli *DNSClient) SetServers(codes, addrs []string, support []bool) {
	cli.lock.Lock()
	cli.servers = completeServers(addrs)
	for i, v := range support {
		cli.servers[i].LocationCode = codes[i]
		cli.servers[i].SupportEDNS = v
	}
	// prom初始化dns服务器数量
	PromDNSOnline.Set(float64(len(addrs)))
	cli.lock.Unlock()
}

// Servers get server state
func (cli *DNSClient) Servers() []*DNSServer {
	return cli.servers
}

// strategy:
// 1. 优先选择第一个服务器
// 2. 每1分钟 dial 一次，如失败则选择下一 dns 服务器。
// 3. 如果所有服务器均失败，选择服务器不变
var (
	selected    int32
	dialTimeout = time.Second * 2
)

var (
	mainlandDomains = []string{
		"baidu.com",
		"tmall.com",
		"163.com",
		"qq.com",
		"jd.com",
	}
	notMainlandDomains = []string{
		"google.com",
		"facebook.com",
		"twitter.com",
		"amazon.com",
		"wikipedia.org",
	}
)

// Ping timer to check dns server state, per minute
func (cli *DNSClient) Ping() {
	chosen := false
	for i, s := range cli.servers {
		domains := mainlandDomains
		if s.LocationCode != "86" {
			domains = notMainlandDomains
		}
		for _, d := range domains {
			ctx := context.WithValue(context.Background(), DnsServerContextKey{}, s.Addr)
			_, err := cli.exchangeOne(ctx, d, dns.TypeA)
			if err != nil {
				if s.State == UP { // 状态转变
					s.State = DOWN
					PromDNSOnline.Dec()
				}
				if cli.callback != nil {
					cli.callback(err, s)
				} else {
					log.Println("dns callback", err, s)
				}
				continue
			}
			if s.State == DOWN { // 状态转变
				s.State = UP
				PromDNSOnline.Inc()
			}

			if s.State == UP {
				// PromDNSState.WithLabelValues(s.Addr).Set(DNSStateUP)
			}

			if !chosen {
				chosen = true
				atomic.CompareAndSwapInt32(&selected, selected, int32(i))
			}
			break
		}
		// 解析所有域名均出错
		if s.State == DOWN {
			// PromDNSState.WithLabelValues(s.Addr).Set(DNSStateDown)
			PromDNSDown.WithLabelValues(s.Addr).Inc()
		}
	}
	time.AfterFunc(time.Minute, cli.Ping)
}

// DNSServerState TODO
type DNSServerState struct {
	Server   string          `json:"server"`
	Loc      string          `json:"loc"`
	StateOK  bool            `json:"state_ok"`
	PingTest map[string]bool `json:"ping_test"`
}

// GetDNSServerStates TODO
func (cli *DNSClient) GetDNSServerStates() []DNSServerState {
	res := make([]DNSServerState, len(cli.servers))
	for i, s := range cli.servers {
		res[i].Server = s.Addr
		res[i].Loc = s.LocationCode
		res[i].PingTest = map[string]bool{}
		domains := mainlandDomains
		if s.LocationCode != "86" {
			domains = notMainlandDomains
		}
		for _, d := range domains {
			res[i].PingTest[d] = true
			ctx := context.WithValue(context.Background(), DnsServerContextKey{}, s.Addr)
			_, err := cli.exchangeOne(ctx, d, dns.TypeA)
			if err != nil {
				res[i].PingTest[d] = false
				continue
			}

			if !res[i].StateOK {
				res[i].StateOK = true
			}
		}
	}
	return res
}

// choose strategy for dns server
func (cli *DNSClient) chooseServer(ctx context.Context) string {
	// specified dns server
	if val, ok := ctx.Value(DnsServerContextKey{}).(string); ok {
		return val
	}
	// edns
	if cli.ednsLocalAddr(ctx) != "" {
		// got the first dns server
		for _, s := range cli.servers {
			if s.SupportEDNS && s.State == UP {
				return s.Addr
			}
		}
	}
	return cli.servers[selected].Addr
}

// use edns
func (cli *DNSClient) ednsLocalAddr(ctx context.Context) string {
	if val, ok := ctx.Value("dns_eaddr").(string); ok {
		return val
	}
	// default not use edns
	return ""
}

// query aaaa
func (cli *DNSClient) allowAAAA(ctx context.Context) bool {
	if val, ok := ctx.Value(DnsAAAAContextKey{}).(bool); ok {
		return val
	}
	// default false
	return false
}

// exchange msg
type dnsResp struct {
	m   *dns.Msg
	err error
}

// the function which for send query message
func (cli *DNSClient) exchangeOne(ctx context.Context, hostname string, qtype uint16) (*dns.Msg, error) {
	m := new(dns.Msg)
	// Set question type
	m.SetQuestion(dns.Fqdn(hostname), qtype)
	// Set DNSSEC OK bit for resolver
	m.SetEdns0(4096, true)
	if len(cli.servers) < 1 {
		return nil, fmt.Errorf("Not configured with at least one DNS Server")
	}
	// Pick a server
	chosenServer := cli.chooseServer(ctx)
	// EDNS Server
	if addr := cli.ednsLocalAddr(ctx); addr != "" {
		ip := net.ParseIP(addr).To4()
		if ip == nil {
			return nil, ErrEmptyEDNSAddr
		}
		e := new(dns.EDNS0_SUBNET)
		e.Code = dns.EDNS0SUBNET
		e.Family = 1         // 1 for IPv4 source address, 2 for IPv6
		e.SourceNetmask = 32 // 32 for IPV4, 128 for IPv6
		e.SourceScope = 0    //
		e.Address = ip       // for IPv4
		for _, v := range m.Extra {
			switch o := v.(type) {
			case *dns.OPT:
				o.Option = append(o.Option, e)
			}
		}
	}
	tries := 1
	for {
		ch := make(chan dnsResp, 1)
		go func() {
			defer utils.Recover(nil)
			rsp, _, err := cli.client.Exchange(m, chosenServer)
			ch <- dnsResp{m: rsp, err: err}
		}()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-ch:
			if r.err != nil {
				operr, ok := r.err.(*net.OpError)
				isRetryable := ok && operr.Temporary()
				hasRetriesLeft := tries < cli.maxTries
				if isRetryable && hasRetriesLeft {
					tries++
					continue
				}
			}
			return r.m, r.err
		}
	}
}

// LookupIPRecord ip records
func LookupIPRecord(ctx context.Context, hostname string) ([]dns.RR, error) {
	rs, err := DefaultClient.LookupIPRecord(ctx, hostname)
	PromCount(err)
	return rs, err
}

// LookupIPRecord TODO
func (cli *DNSClient) LookupIPRecord(ctx context.Context, hostname string) ([]dns.RR, error) {
	var (
		msgA, msgAAAA *dns.Msg
		errA, errAAAA error
	)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer utils.Recover(nil)
		defer wg.Done()
		msgA, errA = cli.exchangeOne(ctx, hostname, dns.TypeA)
		if errA != nil {
			errA = &DNSError{dns.TypeA, hostname, errA, -1}
			return
		}
		if msgA.Rcode != dns.RcodeSuccess {
			errA = &DNSError{dns.TypeA, hostname, nil, msgA.Rcode}
			return
		}
	}()
	if cli.queryAAAA || cli.allowAAAA(ctx) {
		wg.Add(1)
		go func() {
			defer utils.Recover(nil)
			defer wg.Done()
			msgAAAA, errAAAA = cli.exchangeOne(ctx, hostname, dns.TypeAAAA)
			if errAAAA != nil {
				errAAAA = &DNSError{dns.TypeAAAA, hostname, errAAAA, -1}
				return
			}
			if msgAAAA.Rcode != dns.RcodeSuccess {
				errAAAA = &DNSError{dns.TypeAAAA, hostname, nil, msgAAAA.Rcode}
				return
			}
		}()
	}
	wg.Wait()
	if cli.queryAAAA || cli.allowAAAA(ctx) {
		if errA != nil && errAAAA != nil {
			return nil, errA
		}
	} else if errA != nil {
		return nil, errA
	}
	var result []dns.RR
	if msgA != nil {
		result = append(result, msgA.Answer...)
	}
	if msgAAAA != nil {
		result = append(result, msgAAAA.Answer...)
	}
	return result, nil
}

// LookupIP looking for ip, maybe it contain ipv4/ipv6
func LookupIP(ctx context.Context, hostname string) ([]net.IP, error) {
	ips, err := DefaultClient.LookupIP(ctx, hostname)
	PromCount(err)
	return ips, err
}

// LookupIP TODO
func (cli *DNSClient) LookupIP(ctx context.Context, hostname string) ([]net.IP, error) {
	answers, err := cli.LookupIPRecord(ctx, hostname)
	if err != nil {
		return nil, err
	}
	var result []net.IP
	for _, answer := range answers {
		if a, ok := answer.(*dns.A); ok && a.A.To4() != nil && (!IsPrivateV4(a.A) || cli.allowRestrictedAddresses) {
			result = append(result, a.A)
		} else if aaaa, ok := answer.(*dns.AAAA); ok && aaaa.AAAA.To16() != nil && (!IsPrivateV6(aaaa.AAAA) ||
			cli.allowRestrictedAddresses) {
			result = append(result, aaaa.AAAA)
		}
	}
	return result, nil
}

// LookupHost looking for host
func LookupHost(ctx context.Context, hostname string) ([]string, error) {
	hosts, err := DefaultClient.LookupHost(ctx, hostname)
	PromCount(err)
	return hosts, err
}

// LookupHost TODO
func (cli *DNSClient) LookupHost(ctx context.Context, hostname string) ([]string, error) {
	ips, err := cli.LookupIP(ctx, hostname)
	if err != nil {
		return nil, err
	}
	strs := make([]string, len(ips))
	for i, ip := range ips {
		strs[i] = ip.String()
	}
	return strs, nil
}

// LookupMX looking for mx record
func LookupMX(ctx context.Context, hostname string) ([]*dns.MX, error) {
	mxs, err := DefaultClient.LookupMX(ctx, hostname)
	PromCount(err)
	return mxs, err
}

// LookupMX TODO
func (cli *DNSClient) LookupMX(ctx context.Context, hostname string) ([]*dns.MX, error) {
	r, err := cli.exchangeOne(ctx, hostname, dns.TypeMX)
	if err != nil {
		return nil, &DNSError{dns.TypeMX, hostname, err, -1}
	}
	if r.Rcode != dns.RcodeSuccess {
		return nil, &DNSError{dns.TypeMX, hostname, nil, r.Rcode}
	}
	var result []*dns.MX
	for _, answer := range r.Answer {
		if mx, ok := answer.(*dns.MX); ok {
			result = append(result, mx)
		}
	}
	return result, nil
}

// LookupCAA looking for caa record
func LookupCAA(ctx context.Context, hostname string) ([]*dns.CAA, error) {
	caas, err := DefaultClient.LookupCAA(ctx, hostname)
	PromCount(err)
	return caas, err
}

// LookupCAA2 TODO
func (cli *DNSClient) LookupCAA2(ctx context.Context, hostname string) ([]*dns.CAA, error) {
	r, err := cli.exchangeOne(ctx, hostname, dns.TypeCAA)
	if err != nil {
		return nil, &DNSError{dns.TypeCAA, hostname, err, -1}
	}
	if r.Rcode != dns.RcodeSuccess {
		return nil, &DNSError{dns.TypeCAA, hostname, nil, r.Rcode}
	}
	var result []*dns.CAA
	for _, answer := range r.Answer {
		if caa, ok := answer.(*dns.CAA); ok {
			result = append(result, caa)
		}
	}
	return result, nil
}

// rules:
// 1. parse hostname
// 2. parse domains
// 3. recursion domains
var (
	MaxCNAME = 8
	MaxParse = 20
)

// LookupCAA TODO
func (cli *DNSClient) LookupCAA(ctx context.Context, hostname string) ([]*dns.CAA, error) {
	var (
		parseCount   = 0
		subCount     = 0
		currentQuery = hostname
	)

	// 解析域名
	domainName, err := publicsuffix.Parse(hostname)
	if err != nil {
		return nil, err
	}
	if domainName.TRD != "" {
		subCount = strings.Count(domainName.TRD, ".") + 1
	}

	for i := 0; i <= subCount; i++ {
		if i > 0 {
			// 查询子域名
			idx := strings.Index(currentQuery, ".")
			currentQuery = currentQuery[idx+1:]
		}

		var (
			currentCAA = currentQuery
			cnameCount = 0
		)
		for {
			// 限制次数
			if parseCount > MaxParse {
				return nil, ErrTooManyQuery
			}
			if cnameCount > MaxCNAME {
				return nil, ErrTooManyCNAME
			}

			answer, dnsErr := cli.caaAnswer(ctx, currentCAA)
			// 仅返回超时错误，其它错误忽略
			if err != nil && dnsErr.Timeout() {
				dnsErr.hostname = currentCAA
				return nil, dnsErr
			}
			parseCount++

			var tmp []*dns.CAA
			for _, v := range answer {
				if caa, ok := v.(*dns.CAA); ok {
					tmp = append(tmp, caa)
				}
			}
			if len(tmp) > 0 {
				return tmp, nil
			}

			// 解析 cname
			cnames, err := cli.LookupCNAME(ctx, currentCAA)
			if err != nil {
				if dnsErr, ok := err.(*DNSError); ok {
					// ingnore name error and server failure
					if dnsErr.rCode != dns.RcodeNameError &&
						dnsErr.rCode != dns.RcodeServerFailure {
						return nil, err
					}
				}
			}
			parseCount++

			// 没有 CNAME
			if len(cnames) == 0 {
				if domainName.TRD == "" {
					return tmp, nil
				}
				break
			}
			cnameCount++
			// 有 CANAME，去掉最后一个点 .
			currentCAA = cnames[0].Target[:len(cnames[0].Target)-1]
		}
	}
	return nil, nil
}

func (cli *DNSClient) caaAnswer(ctx context.Context, hostname string) ([]dns.RR, *DNSError) {
	r, err := cli.exchangeOne(ctx, hostname, dns.TypeCAA)
	if err != nil {
		return nil, &DNSError{dns.TypeCAA, "", err, -1}
	}
	if r.Rcode != dns.RcodeSuccess {
		return nil, &DNSError{dns.TypeCAA, "", nil, r.Rcode}
	}
	return r.Answer, nil
}

// LookupTXT looking for txt record
func LookupTXT(ctx context.Context, hostname string) ([]*dns.TXT, error) {
	txts, err := DefaultClient.LookupTXT(ctx, hostname)
	PromCount(err)
	return txts, err
}

// LookupTXT TODO
func (cli *DNSClient) LookupTXT(ctx context.Context, hostname string) ([]*dns.TXT, error) {
	r, err := cli.exchangeOne(ctx, hostname, dns.TypeTXT)
	if err != nil {
		return nil, &DNSError{dns.TypeTXT, hostname, err, -1}
	}
	if r.Rcode != dns.RcodeSuccess {
		return nil, &DNSError{dns.TypeTXT, hostname, nil, r.Rcode}
	}
	var result []*dns.TXT
	for _, answer := range r.Answer {
		if answer.Header().Rrtype == dns.TypeTXT {
			if txt, ok := answer.(*dns.TXT); ok {
				result = append(result, txt)
			}
		}
	}
	return result, nil
}

// LookupCNAME looking for cname record
func LookupCNAME(ctx context.Context, hostname string) ([]*dns.CNAME, error) {
	cnames, err := DefaultClient.LookupCNAME(ctx, hostname)
	PromCount(err)
	return cnames, err
}

// LookupCNAME TODO
func (cli *DNSClient) LookupCNAME(ctx context.Context, hostname string) ([]*dns.CNAME, error) {
	r, err := cli.exchangeOne(ctx, hostname, dns.TypeCNAME)
	if err != nil {
		return nil, &DNSError{dns.TypeCNAME, hostname, err, -1}
	}
	if r.Rcode != dns.RcodeSuccess {
		return nil, &DNSError{dns.TypeCNAME, hostname, nil, r.Rcode}
	}
	var result []*dns.CNAME
	for _, answer := range r.Answer {
		if answer.Header().Rrtype == dns.TypeCNAME {
			if cname, ok := answer.(*dns.CNAME); ok {
				result = append(result, cname)
			}
		}
	}
	return result, nil
}

// LookupNS looking for ns record
func LookupNS(ctx context.Context, hostname string) ([]*dns.NS, error) {
	nss, err := DefaultClient.LookupNS(ctx, hostname)
	PromCount(err)
	return nss, err
}

// LookupNS TODO
func (cli *DNSClient) LookupNS(ctx context.Context, hostname string) ([]*dns.NS, error) {
	r, err := cli.exchangeOne(ctx, hostname, dns.TypeNS)
	if err != nil {
		return nil, &DNSError{dns.TypeNS, hostname, err, -1}
	}
	if r.Rcode != dns.RcodeSuccess {
		return nil, &DNSError{dns.TypeNS, hostname, nil, r.Rcode}
	}
	var result []*dns.NS
	for _, answer := range r.Answer {
		if answer.Header().Rrtype == dns.TypeNS {
			if ns, ok := answer.(*dns.NS); ok {
				result = append(result, ns)
			}
		}
	}
	return result, nil
}

// Servers some export functions
func Servers() []*DNSServer {
	return DefaultClient.Servers()
}

// SetServers TODO
func SetServers(codes, addrs []string, support []bool) {
	DefaultClient.SetServers(codes, addrs, support)
}
