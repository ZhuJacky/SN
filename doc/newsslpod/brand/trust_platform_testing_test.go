package brand

import (
	"crypto/x509"
	"testing"
)

var (
	chains = map[string][]*CertInfo{

		// www.baidu.com 交叉证书
		"baidu,across": []*CertInfo{
			&CertInfo{
				SHA1: "D0AE72F9B457343EDD3434EAB2E45F730D78774A",
				Pin:  "5P6dSHqygU3jNqzo8KZ7zW+q5PlcEHk4wubQHgOSNwY=",
			},
			&CertInfo{
				SHA1: "FF67367C5CD4DE4AE18BCCE1D70FDABD7C866135",
				Pin:  "9n0izTnSRF+W4W4JTq51avSXkWhQB8duS2bxVLfzXsY=",
			},
			&CertInfo{
				SHA1: "32F30882622B87CF8856C63DB873DF0853B4DD27",
				Pin:  "JbQbUG5JMJUoI6brnx0x3vZF6jilxsapbXGVfjhN8Fg=",
			},
			&CertInfo{
				SHA1: "A1DB6393916F17E4185509400415C70240B0AE6B",
				Pin:  "sRJBQqWhpaKIGcc1NA7/jJ4vgWj+47oYfyU7waOS1+I=",
			},
		},

		// cfca 证书
		"cfca": []*CertInfo{
			&CertInfo{
				SHA1:    "6A32CEC53AFF211DBD7EE347BEA52873E19A0500",
				Pin:     "UNLaVYztxAzBCgF8GeApLBLhNL3yOT25ewCzL3HHJS0=",
				PubAlgo: x509.RSA,
			},
			&CertInfo{
				SHA1:    "EE41F772ABCDC99A0A3C44281D8406D80D29342A",
				Pin:     "Ub+bKk6h+qrzcElkX07080gHiz8CDNkXagA7GJtPu0I=",
				PubAlgo: x509.RSA,
			},
			&CertInfo{
				SHA1:    "E2B8294B5584AB6B58C290466CAC3FB8398F8483",
				Pin:     "3V7RwJD59EgGG6qUprsRAXVE6e76ogzHFM5sYz9dxik=",
				PubAlgo: x509.RSA,
			},
		},

		"test1": []*CertInfo{
			&CertInfo{
				SHA1:    "aa733942dd9a124b4eb4c218f41e168fcdb710f8",
				Pin:     "N29yKlWU8XyaUE45w8cNsX+Z86tLuSlD4a7gycBpLuI=",
				PubAlgo: x509.ECDSA,
			},
			&CertInfo{
				SHA1:    "6b53c3b358cef368201f8741b9c5aedeea3861fa",
				Pin:     "3kcNJzkUJ1RqMXJzFX4Zxux5WfETK+uL6Viq9lJNn4o=",
				PubAlgo: x509.ECDSA,
			},
			&CertInfo{
				SHA1:    "d4de20d05e66fc53fe1a50882c78db2852cae474",
				Pin:     "Y9mvm0exBk1JoQ57f9Vm28jKo5lFm/woKcVxrYxu80o=",
				PubAlgo: x509.RSA,
			},
		},
	}
)

func TestPlatformTrustTest(t *testing.T) {

	for k, v := range chains {
		res := PlatformTrustTest(v, true)
		for _, r := range res {

			t.Log(r.Platform, r.Pass, k)
			if k == "baidu.across" && !r.Pass {
				t.Fatal("baidu across cert test failed")
			}

			if (r.Platform == Java_8u161 || r.Platform == Java_6u45) && k == "cfca" {
				if r.Pass {
					t.Fatal("cfca cert test failed")
				}
			}
		}
	}
}
