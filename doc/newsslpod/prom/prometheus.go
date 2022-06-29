// Package prom provides ...
package prom

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"mysslee_qcloud/config"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/sirupsen/logrus"
)

const (
	caPEM = `-----BEGIN CERTIFICATE-----
MIIDDzCCAfegAwIBAgIIUwFnuldxtCowDQYJKoZIhvcNAQELBQAwezELMAkGA1UE
BhMCQ04xFzAVBgNVBAoTDktleU1hbmFnZXIub3JnMTEwLwYDVQQLEyhLZXlNYW5h
Z2VyIFRlc3QgUm9vdCAtIEZvciBUZXN0IFVzZSBPbmx5MSAwHgYDVQQDExdLZXlN
YW5hZ2VyIFRlc3QgUm9vdCBDQTAeFw0xOTA5MjkwNTE1NDlaFw0yOTA5MjkwNTE1
NDlaMHoxCzAJBgNVBAYTAkNOMRcwFQYDVQQKEw5LZXlNYW5hZ2VyLm9yZzExMC8G
A1UECxMoS2V5TWFuYWdlciBUZXN0IFJvb3QgLSBGb3IgVGVzdCBVc2UgT25seTEf
MB0GA1UEAxMWS2V5TWFuYWdlciBUZXN0IEVDQyBDQTBZMBMGByqGSM49AgEGCCqG
SM49AwEHA0IABMrzrsTRsbUmfu30krWpBk2Eshr0vwc+dYQJt/oUo3guYTUL+pz4
ZuGIhzZx6zOvpHpH9E0vKti+kFuLUCOlrUijYzBhMA4GA1UdDwEB/wQEAwIBhjAP
BgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBSYU380tFY43dar9FIf2U1HYe814DAf
BgNVHSMEGDAWgBT77ItyGQbTLe7vkBOlfbWodhG1LzANBgkqhkiG9w0BAQsFAAOC
AQEAN233DMIq2ky2Wp3JcPdWF386/PQ4amtIWXgg8iQi9nOWKqAcBuNybC6lbqmZ
83/4HxVIJopZSAHnATBeAyYIecJgsbTlw9aMDBcoXtqL+cUoFFv7sD2bZhqZE/aJ
8UkN3rwY5tIJj0ZjfnoRq7IRttiH1WaTo4rPoVMyCFvj6YOGibsMP5fYzY2i9Nlm
crJO9zgjdjpZjqH2TdB9OZ7RTSKBoRIeIorWc6KDklG81vpOaTj75nFhT7KCN+j1
aHUPXYbavoaZ99yjweDuNuO3QDFN96XNVF0/pmyZF4+rLgk5xmV5FClf4L+L1toh
+wRx7YGsaAldlQnwecsav4nMZw==
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDujCCAqKgAwIBAgIIUjptCVhoCNUwDQYJKoZIhvcNAQELBQAwezELMAkGA1UE
BhMCQ04xFzAVBgNVBAoTDktleU1hbmFnZXIub3JnMTEwLwYDVQQLEyhLZXlNYW5h
Z2VyIFRlc3QgUm9vdCAtIEZvciBUZXN0IFVzZSBPbmx5MSAwHgYDVQQDExdLZXlN
YW5hZ2VyIFRlc3QgUm9vdCBDQTAeFw0xOTA5MjkwNTE1NDlaFw0zOTA5MjkwNTE1
NDlaMHsxCzAJBgNVBAYTAkNOMRcwFQYDVQQKEw5LZXlNYW5hZ2VyLm9yZzExMC8G
A1UECxMoS2V5TWFuYWdlciBUZXN0IFJvb3QgLSBGb3IgVGVzdCBVc2UgT25seTEg
MB4GA1UEAxMXS2V5TWFuYWdlciBUZXN0IFJvb3QgQ0EwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQDn4hIkQFd8gjePLoT6OAkLA5Pr0BJFU4NF9Xh7gYtr
w73/QrEkwuC3PPoOIf7arBrKYXBZXaR5FpHRsPvjuu5z+sUAO4YpF+EBq4zCpmNY
fc2g+1idne2ql/roNd9IbIh4lJsoUTMzpSPtJoOZK3Epy6y9FmuRStJaW982wJQQ
QrUlgxFe4HoRRF388SXu9QKgHMSAXWAHLd1X7nts85IZ1UFJKZu92yukD5OSHKLo
NsAS2TMRolDg+CDN0XmmkEmq8HvSE4KGFlIioJK3aOsAaJt95scLZIhE8S+M2GK/
XJkBoL2wHon8CCGJVVMvJKZKFsX2eGt3zYr6Li1vXrxRAgMBAAGjQjBAMA4GA1Ud
DwEB/wQEAwIBhjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBT77ItyGQbTLe7v
kBOlfbWodhG1LzANBgkqhkiG9w0BAQsFAAOCAQEAEXphVPK/pw0rPIqUgpuDvmfX
9VOi7X8uJeRrt0y2q1yZhtO21q0GQzJtDDBLK/xJJvxRB50kXWVvIhMSwjI2MXLA
x05qzyIRCySxfo5CMJThuJ0W4REYQHpWctsm0s+ryAeWLfKu1S8gOj6EGsnHwK6c
SrnqFHS7XHElvXEW9BZwb8YSUuBSQsNk4G6bI0VZNFmX3CF40INPy8sgXXXe6nWV
4/sjVkQt7oKTGpBgshBLb++oQwW4H7XUeq4I8bLZluDfI6FbSIP1KH84z53XytTF
yIHTDOEawGUr3pCGDMMaeUlZQk2xKnOdnGVNuVgIkb7wxSvU0Px0aVkNJNorEg==
-----END CERTIFICATE-----`
)

var pushers []*push.Pusher

func InitProm(job string) {
	if !config.Conf.Prometheus.Open {
		return
	}

	addr, err := getMyNodeAddr()
	if err != nil {
		panic(err)
	}
	for _, gateway := range config.Conf.Prometheus.Gateway {
		switch {
		case strings.HasPrefix(gateway, "http://"):
			pusher := newHTTPPusher(gateway, job, addr)
			pushers = append(pushers, pusher)
		case strings.HasPrefix(gateway, "https://"):
			pusher, err := newHTTPSPusher(gateway, job, addr)
			if err != nil {
				panic(err)
			}
			pushers = append(pushers, pusher)
		default:
			panic("prom: unrecognized gateway scheme")
		}
	}

	pushToGateway()
}

var myNodeAddr = "" // 本节点的IP

func getMyNodeAddr() (string, error) {
	if myNodeAddr == "" {
		resp, err := http.Get(config.Conf.Prometheus.EchoURL + "/local-ipv4")
		if err != nil {
			err = errors.New("etcd get my node addr err: " + err.Error())
			return "", err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil || len(data) == 0 {
			err = errors.New("etcd get my node addr err: " + err.Error())
			return "", err
		}
		myNodeAddr = string(data)
	}
	return myNodeAddr, nil
}

func pushToGateway() {
	defer time.AfterFunc(time.Second*15, func() {
		pushToGateway()
	})

	for _, pusher := range pushers {
		err := pusher.Push()
		if err != nil {
			logrus.Error("PrometheusPushGateway.Push ", err)
		}
	}

}

func newHTTPPusher(gateway, job, addr string) *push.Pusher {
	return push.New(gateway, job).
		Grouping("instance", addr).
		Gatherer(prometheus.DefaultGatherer)
}

func newHTTPSPusher(gateway, job, addr string) (*push.Pusher, error) {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(caPEM))
	if !ok {
		return nil, errors.New("prom: invalid ca pems")
	}
	cert, err := tls.LoadX509KeyPair(config.Conf.Prometheus.Cert,
		config.Conf.Prometheus.Key)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		RootCAs:            roots,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	// copy from default transport
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       tlsConfig,
	}
	client := &http.Client{Timeout: 5 * time.Second, Transport: tr}

	// pusher
	return push.New(gateway, job).
		Grouping("instance", addr).
		Client(client).
		Gatherer(prometheus.DefaultGatherer), nil
}

// HandlePrometheus prometheus metrics
func HandlePrometheus(c *gin.Context) {
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
