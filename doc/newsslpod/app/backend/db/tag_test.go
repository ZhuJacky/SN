// Package db provides ...
package db

import "testing"

func TestSearchDomainIdsOfAccountByTags(t *testing.T) {
	var tags = []string{
		"1229",
		"1231",
		"1230",
		"1235",
	}
	ids, total, err := SearchDomainIdsOfAccountByTags(1, tags, 0, 100)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ids)
	t.Log(total)
}
