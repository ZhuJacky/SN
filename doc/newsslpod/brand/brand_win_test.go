package brand

import (
	"testing"
)

func TestAutoUpdateWinCert(t *testing.T) {
	test := true
	_testGenPinHook = &test
	hashPinFromCache = map[string]string{}
	//LoadHashPinFromTrustCAsDB()
	AutoUpdateWinCert()

}
