// Package config provides ...
package config

import (
	"os"
	"strings"
)

// CheckerConf the checker config
type CheckerConf struct {
	Listen    int
	BasicAuth map[string]string
	Task      struct {
		FastWorker  int
		FastTimeout int
		FullWorker  int
		FullTimeout int
	}
	LogPath string
}

// InitChecker init checker config
func InitChecker() {
	basicAuth := os.Getenv("MYSSL_CHECKER_BASICAUTH")
	if basicAuth != "" {
		sli := strings.Split(basicAuth, ",")

		if Conf.Checker.BasicAuth == nil {
			Conf.Checker.BasicAuth = make(map[string]string, len(sli))
		}
		for _, v := range sli {
			up := strings.Split(v, ":")
			Conf.Checker.BasicAuth[up[0]] = up[1]
		}
	}
}
