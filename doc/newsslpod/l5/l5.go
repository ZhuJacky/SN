// Package l5 provides ...
package l5

import "git.code.oa.com/components/l5"

var L5API *l5.Api

func Init() {
	var err error
	L5API, err = l5.NewDefaultApi()
	if err != nil {
		panic(err)
	}
}

// GetRedisServer
func GetRedisServer() (*l5.Server, error) {
	domain, err := L5API.Query("redis")
	if err != nil {
		return nil, err
	}
	return L5API.GetServerBySid(domain.Mod, domain.Cmd)
}
