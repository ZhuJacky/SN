package l5

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type Domain struct {
	sync.Mutex
	balancer Balancer
	api      *Api
	Name     string
	Mod      int32
	Cmd      int32
	expire   time.Time
}

func (d *Domain) Get() (*Server, error) {
	d.Lock()
	defer d.Unlock()
	var (
		srv       *Server
		lastError error
	)

	// step1：从balancer中获取
	if srv, lastError = d.getFromBalancer(); lastError == nil {
		return srv, nil
	}

	// step2：从agent中获取
	if lastError = d.setServersFromAgent(); lastError == nil {
		if srv, lastError = d.getFromBalancer(); lastError == nil {
			return srv, nil
		}
	}

	// step3：从静态文件中获取
	if lastError = d.setServersFromStatic(); lastError == nil {
		if srv, lastError = d.getFromBalancer(); lastError == nil {
			return srv, nil
		}
	}

	return nil, lastError
}

// GetAll servers
func (d *Domain) GetAll() (ss []*Server, err error) {
	d.Lock()
	defer d.Unlock()

	// step1：从balancer中获取
	if ss, err = d.getAllFromBalancer(); err == nil {
		return
	}

	// step2：从agent中获取
	if err = d.setServersFromAgent(); err == nil {
		if ss, err = d.getAllFromBalancer(); err == nil {
			return
		}
	}

	// step3：从静态文件中获取
	if err = d.setServersFromStatic(); err == nil {
		if ss, err = d.getAllFromBalancer(); err == nil {
			return
		}
	}
	return
}

func (d *Domain) getAllFromBalancer() ([]*Server, error) {
	if d.balancer == nil {
		return nil, ErrNotBalancer
	}
	srvAll, err := d.balancer.GetAll()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	ss := make([]*Server, 0, len(srvAll))
	for _, s := range srvAll {
		if s.expire.Before(now) {
			if d.balancer != nil {
				d.balancer.Remove(s)
			}
		} else {
			ss = append(ss, s)
		}
	}

	if len(ss) < 1 {
		return nil, ErrNotFound
	}
	return ss, nil
}

func (d *Domain) getFromBalancer() (*Server, error) {
	if d.balancer == nil {
		return nil, ErrNotBalancer
	}
	srv, err := d.balancer.Get()
	if err != nil {
		return nil, err
	}
	if srv.expire.Before(time.Now()) {
		if d.balancer != nil {
			d.balancer.Remove(srv)
		}
		return nil, ErrNotFound
	}
	return srv.allocate(), nil
}

func (d *Domain) setServersFromAgent() error {
	var (
		now = time.Now()
		cmd = QOS_CMD_BATCH_GET_ROUTE_VER
	)
	if d.api.Balancer == CL5_LB_TYPE_CST_HASH {
		cmd = QOS_CMD_BATCH_GET_ROUTE_WEIGHT
	}
	buf, err := d.api.dial(int32(cmd), d.Mod, d.Cmd, int32(os.Getpid()), int32(Version))
	if err != nil {
		return err
	}
	size := len(buf) - 16
	list := make([]*Server, size/14)
	for k := range list {
		if cmd == QOS_CMD_BATCH_GET_ROUTE_WEIGHT {
			list[k] = &Server{
				domain: d,
				Ip:     net.IPv4(buf[16+k*14], buf[17+k*14], buf[18+k*14], buf[19+k*14]),
				Port:   d.api.Endian.Uint16(buf[20+k*14 : 22+k*14]),
				weight: int32(d.api.Endian.Uint32(buf[22+k*14 : 26+k*14])),
				total:  int32(d.api.Endian.Uint32(buf[26+k*14 : 30+k*14])),
				expire: now.Add(d.api.ServerExpire),
			}
		} else {
			list[k] = &Server{
				domain: d,
				Ip:     net.IPv4(buf[16+k*14], buf[17+k*14], buf[18+k*14], buf[19+k*14]),
				Port:   d.api.Endian.Uint16(buf[20+k*14 : 22+k*14]),
				weight: int32(d.api.Endian.Uint32(buf[22+k*14 : 26+k*14])),
				total:  int32(d.api.Endian.Uint32(buf[22+k*14 : 26+k*14])),
				expire: now.Add(d.api.ServerExpire),
			}
		}
	}
	if len(list) < 1 {
		return ErrNotFound
	}
	err = d.unsafeSet(list)
	if err != nil {
		return err
	}
	return nil
}

func (d *Domain) setServersFromStatic() error {
	var (
		fp  *os.File
		err error
		now = time.Now()
	)
	list := make([]*Server, 0)
	for _, v := range d.api.StaticServerFiles {
		if fp, err = os.Open(v); err != nil {
			continue
		}
		for {
			var (
				mod  int32
				cmd  int32
				ip   string
				port uint16
			)
			if n, fail := fmt.Fscanln(fp, &mod, &cmd, &ip, &port); n == 0 || fail != nil {
				break
			}
			if d.Mod != mod || d.Cmd != cmd {
				continue
			}
			list = append(list, &Server{
				domain: d,
				Ip:     net.ParseIP(ip),
				Port:   port,
				weight: 100, // default weight: 100
				total:  0,
				expire: now.Add(d.api.ServerExpire),
			})
		}
		fp.Close()
	}

	if len(list) < 1 {
		return ErrNotFound
	}

	err = d.unsafeSet(list)
	if err != nil {
		return err
	}
	return nil
}

func (d *Domain) unsafeSet(list []*Server) error {
	if d.balancer == nil {
		return ErrNotBalancer
	}
	if err := d.balancer.Destroy(); err != nil {
		return err
	}
	for _, v := range list {
		if err := d.balancer.Set(v); err != nil {
			return err
		}
	}
	return nil
}

func (d *Domain) Set(list []*Server) error {
	d.Lock()
	defer d.Unlock()
	return d.unsafeSet(list)
}

func (d *Domain) unsafeDestroy() error {
	if d.balancer == nil {
		return ErrNotBalancer
	}
	return d.balancer.Destroy()
}

func (d *Domain) Destroy() error {
	d.Lock()
	defer d.Unlock()
	return d.unsafeDestroy()
}

func (d *Domain) SetBalancer(b Balancer) {
	d.Lock()
	defer d.Unlock()
	d.balancer = b
}

type Domains map[string]*Domain

type SidCache map[string]*Domain

//根据l5名称查询l5 mod+cmd
func (c *Api) Query(name string) (*Domain, error) {
	now := time.Now()
	c.RLock()
	domain, exists := c.Domains[name]
	c.RUnlock()
	if exists && domain.expire.After(now) {
		return domain, nil
	} else {
		domain = &Domain{api: c, Name: name, Mod: 0, Cmd: 0, expire: now.Add(c.DomainExpire), balancer: NewBalancer(c.Balancer)}
	}
	buf, err := c.dial(QOS_CMD_QUERY_SNAME, domain.Mod, domain.Cmd, int32(os.Getpid()), int32(len(domain.Name)), domain.Name)
	if err != nil {
		return nil, err
	}
	domain.Mod = int32(c.Endian.Uint32(buf[0:4]))
	domain.Cmd = int32(c.Endian.Uint32(buf[4:8]))
	c.Lock()
	c.Domains[name] = domain
	c.Unlock()
	return domain, nil
}

//定时load静态domain文件
func (c *Api) interval() {
	interval := time.NewTicker(c.StaticDomainReload)
	for {
		select {
		case <-interval.C:
			var (
				err error
				fp  *os.File
			)
			for _, v := range c.StaticDomainFiles {
				if fp, err = os.Open(v); err != nil {
					//log.Printf("open file failed: %s", err.Error())
					continue
				}
				for {
					var (
						name string
						mod  int32
						cmd  int32
					)
					if n, fail := fmt.Fscanln(fp, &name, &mod, &cmd); n == 0 || fail != nil {
						break
					}
					now := time.Now()
					c.Lock()
					_, exists := c.Domains[name]
					if !exists {
						c.Domains[name] = &Domain{}
					}
					if c.Domains[name].expire.Before(now) {
						c.Domains[name].Lock()
						c.Domains[name].api = c
						c.Domains[name].balancer = NewBalancer(c.Balancer)
						c.Domains[name].Cmd = cmd
						c.Domains[name].Mod = mod
						c.Domains[name].expire = now.Add(c.DomainExpire)
						c.Domains[name].Unlock()
					}
					c.Unlock()
				}
				fp.Close()
			}
		}
	}
}
