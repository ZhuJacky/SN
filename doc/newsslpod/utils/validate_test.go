// Package utils provides ...
package utils

import (
	"io/ioutil"
	"strings"
	"testing"

	"golang.org/x/net/idna"
)

func TestVadlidateDomain(t *testing.T) {
	var tests = []string{
		"www._test.com",
	}
	for _, v := range tests {
		ok := ValidateDomain(v)
		if !ok {
			t.Fatal(v)
		}
	}
}

func TestValidateDomain2(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/domains.txt")
	if err != nil {
		t.Fatal(err)
	}
	domains := string(data)
	for i, line := range strings.Split(domains, "\n") {
		if line == "" {
			continue
		}
		punycode, err := idna.ToASCII(strings.TrimSpace(line))
		if err != nil {
			t.Fatal(i, line, err)
		}
		ok := ValidateDomain2(punycode)
		if !ok {
			t.Fatal(i, line, ok)
		}
	}

	tests := []string{
		"-baidu.com",
		"eabaidu-.com",
		"1234567890123456789012345678901234567890123456789012345678901234.com",
		"abc.efadsfasfd.is.fffadsfadsfail",
	}
	for _, v := range tests {
		ok := ValidateDomain2(v)
		if ok {
			t.Fatal(v, ok)
		}
	}
}
