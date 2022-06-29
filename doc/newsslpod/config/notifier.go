// Package config provides ...
package config

// NotifierConf the notifier config
type NotifierConf struct {
	Listen    int
	Domain    string
	LogPath   string
	RateLimit int
	// Phone  phone
	// Email  email
	// Wechat wechat
}

type phone struct {
	API  string
	User string
	Key  string
}

type email struct {
	User     string
	Key      string
	API      string
	From     string
	FromName string
}

type wechat struct {
}

// InitNotifier init notifier config
func InitNotifier() {
	// TODO
}
