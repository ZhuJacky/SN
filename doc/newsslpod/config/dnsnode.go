// Package config provides ...
package config

import "strings"

const (
	NodeTypeCN = iota
	NodeTypeHK
	NodeTypeUS
)

type DNSNode struct {
	LocationCode string `json:"location_code"`  // node code with location, use country's code and city's code
	Addr         string `json:"addr"`           // ip/host address
	Name         string `json:"name,omitempty"` // node name
	SupportEDNS  bool   `json:"support_edns"`   // support Edns?
	Level        int    `json:"level"`          // priority
}

func (n DNSNode) IsZero() bool {
	return n.LocationCode == ""
}

type DNSNodes []DNSNode

// 获取 addr 地址
func (ns DNSNodes) Infos() (codes, addrs []string, support []bool) {
	if ns == nil {
		return nil, nil, nil
	}
	codes = make([]string, len(ns))
	addrs = make([]string, len(ns))
	support = make([]bool, len(ns))
	for i, v := range ns {
		codes[i] = v.LocationCode
		addrs[i] = v.Addr
		support[i] = v.SupportEDNS
	}
	return
}

// 获取
func (ns DNSNodes) GetNode(addr string) DNSNode {
	for _, node := range ns {
		if strings.Contains(addr, node.Addr) {
			return node
		}
	}
	return DNSNode{}
}

// length
func (ns DNSNodes) Len() int {
	return len(ns)
}

// city data
type Location struct {
	Code  string     `json:"code"`
	Name  string     `json:"name"`
	Child []Location `json:"child,omitempty"`
}
