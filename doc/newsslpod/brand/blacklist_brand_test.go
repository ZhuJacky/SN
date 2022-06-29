package brand

import (
	"crypto/x509"
	"testing"

	"mysslee_qcloud/utils/certutils"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	WoSignRoot1 = `-----BEGIN CERTIFICATE-----
MIIFdjCCA16gAwIBAgIQXmjWEXGUY1BWAGjzPsnFkTANBgkqhkiG9w0BAQUFADBV
MQswCQYDVQQGEwJDTjEaMBgGA1UEChMRV29TaWduIENBIExpbWl0ZWQxKjAoBgNV
BAMTIUNlcnRpZmljYXRpb24gQXV0aG9yaXR5IG9mIFdvU2lnbjAeFw0wOTA4MDgw
MTAwMDFaFw0zOTA4MDgwMTAwMDFaMFUxCzAJBgNVBAYTAkNOMRowGAYDVQQKExFX
b1NpZ24gQ0EgTGltaXRlZDEqMCgGA1UEAxMhQ2VydGlmaWNhdGlvbiBBdXRob3Jp
dHkgb2YgV29TaWduMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAvcqN
rLiRFVaXe2tcesLea9mhsMMQI/qnobLMMfo+2aYpbxY94Gv4uEBf2zmoAHqLoE1U
fcIiePyOCbiohdfMlZdLdNiefvAA5A6JrkkoRBoQmTIPJYhTpA2zDxIIFgsDcScc
f+Hb0v1naMQFXQoOXXDX2JegvFNBmpGN9J42Znp+VsGQX+axaCA2pIwkLCxHC1l2
ZjC1vt7tj/id07sBMOby8w7gLJKA84X5KIq0VC6a7fd2/BVoFutKbOsuEo/Uz/4M
x1wdC34FMr5esAkqQtXJTpCzWQ27en7N1QhatH/YHGkR+ScPewavVIMYe+HdVHpR
aG53/Ma/UkpmRqGyZxq7o093oL5d//xWC0Nyd5DKnvnyOfUNqfTq1+ezEC8wQjch
zDBwyYaYD8xYTYO7feUapTeNtqwylwA6Y3EkHp43xP901DfA4v6IRmAR3Qg/UDar
uHqklWJqbrDKaiFaafPz+x1wOZXzp26mgYmhiMU7ccqjUu6Du/2gd/Tkb+dC221K
mYo0SLwX3OSACCK28jHAPwQ+658geda4BmRkAjHXqc1S+4RFaQkAKtxVi8QGRkvA
Sh0JWzko/amrzgD5LkhLJuYwTKVYyrREgk/nkR4zw7CT/xH8gdLKH3Ep3XZPkiWv
HYG3Dy+MwwbMLyejSuQOmbp8HkUff6oZRZb9/D0CAwEAAaNCMEAwDgYDVR0PAQH/
BAQDAgEGMA8GA1UdEwEB/wQFMAMBAf8wHQYDVR0OBBYEFOFmzw7R8bNLtwYgFP6H
EtX2/vs+MA0GCSqGSIb3DQEBBQUAA4ICAQCoy3JAsnbBfnv8rWTjMnvMPLZdRtP1
LOJwXcgu2AZ9mNELIaCJWSQBnfmvCX0KI4I01fx8cpm5o9dU9OpScA7F9dY74ToJ
MuYhOZO9sxXqT2r09Ys/L3yNWC7F4TmgPsc9SnOeQHrAK2GpZ8nzJLmzbVUsWh2e
JXLOC62qx1ViC777Y7NhRCOjy+EaDveaBk3e1CNOIZZbOVtXHS9dCF4Jef98l7VN
g64N1uajeeAz0JmWAjCnPv/So0M/BVoG6kQC2nz4SNAzqfkHx5Xh9T71XXG68pWp
dIhhWeO/yloTunK0jF02h+mmxTwTv97QRCbut+wucPrXnbes5cVAWubXbHssw1ab
R80LzvobtCHXt2a49CUwi1wNuepnsvRtrtWhnk/Yn+knArAdBtaP4/tIEp9/EaEQ
PkxROpaw0RPxx9gmrjrKkcRpnd8BKWRRb2jaFOwIQZeQjdCygPLPwj2/kWjFgGce
xGATVdVhmVd8upUPYUk6ynW8yQqTP2cOEvIo4jEbwFcW3wh8GcF+Dx+FHgo2fFt+
J7x6v+Db9NpSvd4MVHAxkUOVyLzwPt0JfjBkUO1/AaQzZ01oT74V77D2AhGiGxMl
OtzCWfHjXEa7ZywCRuoeSKbmW9m1vFGikpbbqsY3Iqb+zCB0oy2pLmvLwIIRIbWT
ee5Ehr7XHuQe+w==
-----END CERTIFICATE-----`
	wosignAndStartCom = `-----BEGIN CERTIFICATE-----
MIIDfDCCAmSgAwIBAgIQayXaioidfLwPBbOxemFFRDANBgkqhkiG9w0BAQsFADBY
MQswCQYDVQQGEwJDTjEaMBgGA1UEChMRV29TaWduIENBIExpbWl0ZWQxLTArBgNV
BAMTJENlcnRpZmljYXRpb24gQXV0aG9yaXR5IG9mIFdvU2lnbiBHMjAeFw0xNDEx
MDgwMDU4NThaFw00NDExMDgwMDU4NThaMFgxCzAJBgNVBAYTAkNOMRowGAYDVQQK
ExFXb1NpZ24gQ0EgTGltaXRlZDEtMCsGA1UEAxMkQ2VydGlmaWNhdGlvbiBBdXRo
b3JpdHkgb2YgV29TaWduIEcyMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKC
AQEAvsXEoCKASU+/2YcRxlPhuw+9YH+v9oIOH9ywjj2X4FA8jzrvZjtFB5sg+OPX
JYY1kBaiXW8wGQiHC38Gsp1ij96vkqVg1CuAmlI/9ZqD6TRay9nVYlzmDuDfBpgO
gHzKtB0TiGsOqCR3A9DuW/PKaZE1OVbFbeP3PU9ekzgkyhjpJMuSA93MHD0JcOQg
5PGurLtzaaNjOg9FD6FKmsLRY6zLEPg95k4ot+vElbGs/V6r+kHLXZ1L3PR8du9n
fwB6jdKgGlxNIuG12t12s9R23164i5jIFFTMaxeSt+BKv0mUYQs4kI9dJGwlezt5
2eJ+na2fmKEG/HgUYFf47oB3sQIDAQABo0IwQDAOBgNVHQ8BAf8EBAMCAQYwDwYD
VR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQU+mCp62XF3RYUCE4MD42b4Pdkr2cwDQYJ
KoZIhvcNAQELBQADggEBAFfDejaCnI2Y4qtAqkePx6db7XznPWZaOzG73/MWM5H8
fHulwqZm46qwtyeYP0nXYGdnPzZPSsvxFPpahygc7Y9BMsaV+X3avXtbwrAh449G
3CE4Q3RM+zD4F3LBMvzIkRfEzFg3TgvMWvchNSiDbGAtROtSjFA9tWwS1/oJu2yy
SrHFieT801LYYRf+epSEj3m2M1m6D8QL4nCgS3gu+sif/a+RZQp4OBXllxcU3fng
LDT4ONCEIgDAFFEYKwLcMFrw6AF8NTojrwjkr6qOKEJJLvD1mTS+7Q9LGOHSJDy7
XUe3IfKN0QqZjuNuPq1w4I+5ysxugTH2e5x6eeRncRg=
-----END CERTIFICATE-----`
	letsEncrypt = `-----BEGIN CERTIFICATE-----
MIIEkjCCA3qgAwIBAgIQCgFBQgAAAVOFc2oLheynCDANBgkqhkiG9w0BAQsFADA/
MSQwIgYDVQQKExtEaWdpdGFsIFNpZ25hdHVyZSBUcnVzdCBDby4xFzAVBgNVBAMT
DkRTVCBSb290IENBIFgzMB4XDTE2MDMxNzE2NDA0NloXDTIxMDMxNzE2NDA0Nlow
SjELMAkGA1UEBhMCVVMxFjAUBgNVBAoTDUxldCdzIEVuY3J5cHQxIzAhBgNVBAMT
GkxldCdzIEVuY3J5cHQgQXV0aG9yaXR5IFgzMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAnNMM8FrlLke3cl03g7NoYzDq1zUmGSXhvb418XCSL7e4S0EF
q6meNQhY7LEqxGiHC6PjdeTm86dicbp5gWAf15Gan/PQeGdxyGkOlZHP/uaZ6WA8
SMx+yk13EiSdRxta67nsHjcAHJyse6cF6s5K671B5TaYucv9bTyWaN8jKkKQDIZ0
Z8h/pZq4UmEUEz9l6YKHy9v6Dlb2honzhT+Xhq+w3Brvaw2VFn3EK6BlspkENnWA
a6xK8xuQSXgvopZPKiAlKQTGdMDQMc2PMTiVFrqoM7hD8bEfwzB/onkxEz0tNvjj
/PIzark5McWvxI0NHWQWM6r6hCm21AvA2H3DkwIDAQABo4IBfTCCAXkwEgYDVR0T
AQH/BAgwBgEB/wIBADAOBgNVHQ8BAf8EBAMCAYYwfwYIKwYBBQUHAQEEczBxMDIG
CCsGAQUFBzABhiZodHRwOi8vaXNyZy50cnVzdGlkLm9jc3AuaWRlbnRydXN0LmNv
bTA7BggrBgEFBQcwAoYvaHR0cDovL2FwcHMuaWRlbnRydXN0LmNvbS9yb290cy9k
c3Ryb290Y2F4My5wN2MwHwYDVR0jBBgwFoAUxKexpHsscfrb4UuQdf/EFWCFiRAw
VAYDVR0gBE0wSzAIBgZngQwBAgEwPwYLKwYBBAGC3xMBAQEwMDAuBggrBgEFBQcC
ARYiaHR0cDovL2Nwcy5yb290LXgxLmxldHNlbmNyeXB0Lm9yZzA8BgNVHR8ENTAz
MDGgL6AthitodHRwOi8vY3JsLmlkZW50cnVzdC5jb20vRFNUUk9PVENBWDNDUkwu
Y3JsMB0GA1UdDgQWBBSoSmpjBH3duubRObemRWXv86jsoTANBgkqhkiG9w0BAQsF
AAOCAQEA3TPXEfNjWDjdGBX7CVW+dla5cEilaUcne8IkCJLxWh9KEik3JHRRHGJo
uM2VcGfl96S8TihRzZvoroed6ti6WqEBmtzw3Wodatg+VyOeph4EYpr/1wXKtx8/
wApIvJSwtmVi4MFU5aMqrSDE6ea73Mj2tcMyo5jMd6jmeWUHK8so/joWUoHOUgwu
X4Po1QYz+3dszkDqMp4fklxBwXRsW10KXzPMTZ+sOPAveyxindmjkW8lGy+QsRlG
PfZ+G6Z6h7mjem0Y+iWlkYcV4PIWL1iwBi8saCbGS5jN2p8M+X+Q7UNKEkROb3N6
KOqkqm57TH2H3eDJAkSnh6/DNFu0Qg==
-----END CERTIFICATE-----
`
)

//判断证书是否在黑名单中
func TestIsInBlackList(t *testing.T) {
	assert := assert.New(t)

	certs, ok := certutils.GenCertsFromPem([]byte(wosignAndStartCom))
	if !ok {
		log.Panicf("加载wosignAndStartCom")
	}
	wosignAndStartComCert := certs[0]

	certs, ok = certutils.GenCertsFromPem([]byte(letsEncrypt))
	if !ok {
		log.Panicf("加载cert")
	}
	letsEncryptCert := certs[0]

	tests := []struct {
		cert   *x509.Certificate
		result bool
	}{
		{wosignAndStartComCert, true},
		{letsEncryptCert, false},
	}

	for _, test := range tests {
		result := IsInBlackList(certutils.GenPin(test.cert.RawSubjectPublicKeyInfo))
		assert.Equal(test.result, result)
	}

}

//测试是否在品牌黑名单中
func TestIsInBlackBrand(t *testing.T) {
	assert := assert.New(t)
	InitBlack()
	certs, ok := certutils.GenCertsFromPem([]byte(WoSignRoot1))
	if !ok {
		log.Panic("")
	}
	woSign := certs[0]

	tests := []struct {
		cert   *x509.Certificate
		result bool
	}{
		{woSign, true},
	}

	for _, test := range tests {
		result := IsInBlackBrand(test.cert.Issuer)
		assert.Equal(test.result, result)
	}
}
