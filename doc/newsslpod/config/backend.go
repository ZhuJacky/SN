// Package config provides ...
package config

import (
	"os"
	"strconv"
	"strings"
)

// BackendConf the backend config
type BackendConf struct {
	Listen    int
	BasicAuth map[string]string
	BasicPlan int
	Task      struct {
		Duration float64
		Interval int
	}
	LogPath string
}

// InitBackend init backend config
func InitBackend() {
	port := os.Getenv("MYSSLEE_BACKEND_PORT")
	if port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			panic("MYSSLEE_BACKEND_PORT should be int: " + port)
		}
		Conf.Backend.Listen = p
	}
	basicAuth := os.Getenv("MYSSLEE_BACKEND_BASICAUTH")
	if basicAuth != "" {
		sli := strings.Split(basicAuth, ",")

		if Conf.Backend.BasicAuth == nil {
			Conf.Backend.BasicAuth = make(map[string]string, len(sli))
		}
		for _, v := range sli {
			up := strings.Split(v, ":")
			Conf.Backend.BasicAuth[up[0]] = up[1]
		}
	}
}
