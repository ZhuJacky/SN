package l5

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

type Option struct {
	Endian                  binary.ByteOrder
	Host                    string
	Port                    int16
	Timeout                 time.Duration
	MaxConn                 int
	MaxPacketSize           int
	StaticDomainFiles       []string
	DomainExpire            time.Duration
	StaticDomainReload      time.Duration
	StaticServerFiles       []string
	ServerExpire            time.Duration
	StatErrorReportInterval time.Duration
	StatReportInterval      time.Duration
	StatMaxErrorCount       int
	StatMaxErrorRate        float64
	Balancer                int
}

type Api struct {
	sync.RWMutex
	*Option
	Domains
	SidCache
	seq  uint32
	conn chan net.Conn
}

func NewApi(opt *Option) (*Api, error) {
	api := &Api{
		RWMutex:  sync.RWMutex{},
		Option:   opt,
		Domains:  make(Domains),
		SidCache: make(SidCache),
		conn:     make(chan net.Conn, opt.MaxConn),
		seq:      0,
	}
	for i := 0; i < opt.MaxConn; i++ {
		conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", opt.Host, opt.Port), opt.Timeout)
		if err != nil {
			return nil, err
		}
		api.conn <- conn
	}
	go api.interval()
	return api, nil
}

func NewDefaultApi() (*Api, error) {
	return NewApi(&Option{
		Endian:                  binary.LittleEndian,                                                                       //默认小端序
		Host:                    "127.0.0.1",                                                                               //默认agent ip
		Port:                    8888,                                                                                      //默认agent port
		Timeout:                 time.Second,                                                                               //操作超时
		MaxConn:                 8,                                                                                         //连接池大小
		MaxPacketSize:           1024,                                                                                      //agent通信允许最大包
		StaticDomainFiles:       []string{"/data/L5Backup/name2sid.backup", "/data/L5Backup/name2sid.cache.bin"},           //默认domain静态文件
		DomainExpire:            30 * time.Second,                                                                          //domain有效期
		StaticDomainReload:      30 * time.Second,                                                                          //静态domain重载时间
		StaticServerFiles:       []string{"/data/L5Backup/current_route.backup", "/data/L5Backup/current_route_v2.backup"}, //默认server静态文件
		ServerExpire:            30 * time.Second,                                                                          //server有效期
		StatErrorReportInterval: time.Second,                                                                               //错误上报间隔
		StatReportInterval:      5 * time.Second,                                                                           //正常上报间隔
		StatMaxErrorCount:       16,                                                                                        //最大错误数
		StatMaxErrorRate:        0.2,                                                                                       //最大错误比例
		Balancer:                DefaultBalancer,                                                                           //balancer类型
	})
}

//通过l5名称获取服务器信息
func (c *Api) GetServerByName(name string) (*Server, error) {
	domain, err := c.Query(name)
	if err != nil {
		return nil, err
	}
	return domain.Get()
}

//通过sid获取服务器信息
func (c *Api) GetServerBySid(mod int32, cmd int32) (*Server, error) {
	sid := fmt.Sprintf("%d_%d", mod, cmd)
	c.RLock()
	if domain, exists := c.SidCache[sid]; exists {
		c.RUnlock()
		return domain.Get()
	}
	c.RUnlock()
	c.Lock()
	defer c.Unlock()
	c.SidCache[sid] = &Domain{api: c, Mod: mod, Cmd: cmd, expire: time.Now().Add(c.DomainExpire), balancer: NewBalancer(c.Balancer)}
	return c.SidCache[sid].Get()
}

// GetServers 返回sid关联的所有服务器信息列表
func (c *Api) GetServersBySid(mod int32, cmd int32) ([]*Server, error) {
	sid := fmt.Sprintf("%d_%d", mod, cmd)
	c.RLock()
	if domain, exists := c.SidCache[sid]; exists {
		c.RUnlock()
		return domain.GetAll()
	}
	c.RUnlock()
	c.Lock()
	defer c.Unlock()
	c.SidCache[sid] = &Domain{api: c, Mod: mod, Cmd: cmd, expire: time.Now().Add(c.DomainExpire), balancer: NewBalancer(c.Balancer)}
	return c.SidCache[sid].GetAll()
}
