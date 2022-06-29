// Package email provides ...
package email

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"mysslee_qcloud/config"

	log "github.com/sirupsen/logrus"
)

type sohuResult struct {
	StatusCode int
	Info       map[string]interface{}
	Message    string
	Result     bool
}

func SendMail(subject string, to []string, msg []byte, from string) error {
	log.Println("mysslee.SendMail send to ", to)

	vals := url.Values{}
	vals.Set("apiUser", config.Conf.Notifier.Email.User)
	vals.Set("apiKey", config.Conf.Notifier.Email.Key)
	if from != "" {
		vals.Set("from", from)
	} else {
		vals.Set("from", config.Conf.Notifier.Email.From)
	}
	vals.Set("to", strings.Join(to, ";"))
	vals.Set("subject", subject)
	vals.Set("html", string(msg))
	vals.Set("fromName", config.Conf.Notifier.Email.FromName)

	client := http.Client{}
	response, err := client.PostForm(config.Conf.Notifier.Email.API, vals)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	var rst sohuResult
	err = json.Unmarshal(data, &rst)
	if err != nil {
		return err
	}

	if !rst.Result || rst.StatusCode != 200 {
		return fmt.Errorf("%s,%v", rst.Message, rst.Info)
	}
	return nil
}
