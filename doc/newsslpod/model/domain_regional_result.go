package model

import "time"

// DomainRegionalResult 多地域检测结果表
type DomainRegionalResult struct {
	DomainID          int       `gorm:"type:int;primary_key;not null"`
	Region            string    `gorm:"type:varchar(64);primary_key;not null"`
	DetectionResult   string    `gorm:"type:text"`
	ResultHash        string    `gorm:"type:varchar(64)"`
	LastDetectionTime time.Time `gorm:"column:last_detection_time;index"`
}

// CallbackToBackend checker发给backend的回调 ip:证书
type CallbackToBackend struct {
	IPCerts map[string]CertWithErr
	Err     string
}

// CertWithErr 带有错误信息的证书
type CertWithErr struct {
	Cert   []*CertInfo
	IPorts IPPort
	Err    string
}

// DetectionResult 详细检测信息 ip-status-hash
type DetectionResult struct {
	IP     string
	Port   string
	Status string
	Hashes []string
	IsAuto bool
}

// KafkaDomainInfo 需要推送到kafka的域名信息
type KafkaDomainInfo struct {
	DomainID       int      `json:"domain_id"`
	Domain         string   `json:"domain"`
	PunyCodeDomain string   `json:"punycode_doman"`
	ServerType     int      `json:"server_type"`
	IPPorts        []IPPort `json:"ip_ports"`
	IsAutoDetect   bool     `json:"is_auto_detect"`
	DomainFlag     int      `json:"domain_flag"`
	IP             string   `json:"ip"`
	Port           string   `json:"port"`
}

// IPPort 域名端口
type IPPort struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}
