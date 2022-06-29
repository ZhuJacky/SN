// Package redis provides ...
package redis

import (
	"errors"
	"io/ioutil"
	"net/http"

	"mysslee_qcloud/config"
)

type AppName string

const (
	BackendApp  AppName = "app:backend:"
	CheckerApp  AppName = "app:checker:"
	NotifierApp AppName = "app:notifier:"

	suffixInstallType = "/instance-type"
	suffixLocalIPV4   = "/local-ipv4"
)

var selfIP string

func selfIPAddress() (string, error) {
	if selfIP == "" {
		resp, err := http.Get(config.Conf.Prometheus.EchoURL + suffixLocalIPV4)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode/100 != 2 {
			return "", errors.New(resp.Status)
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil || len(data) == 0 {
			return "", errors.New("selfIPAddress.ReadAll err")
		}
		selfIP = string(data)
	}
	return selfIP, nil
}
