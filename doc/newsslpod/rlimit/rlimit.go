// Package limit provides ...
package limit

import (
	"sync"
	"time"
)

var LimitMap *rateLimitMgr

func init() {
	LimitMap = &rateLimitMgr{
		lock: sync.RWMutex{},
		maps: make(map[string]*rateLimit),
	}
	go LimitMap.start()
}

type rateLimitMgr struct {
	lock sync.RWMutex
	maps map[string]*rateLimit
}

func (mgr *rateLimitMgr) start() {
	// clean
	t := time.NewTicker(time.Second * 59)
	for now := range t.C {
		mgr.lock.Lock()
		for k, v := range mgr.maps {
			if now.Add(-v.rate).Unix() > v.Reset {
				v = nil
				delete(mgr.maps, k)
			}
		}
		mgr.lock.Unlock()
	}
}

func (mgr *rateLimitMgr) Load(key string, limit int, rate time.Duration) *rateLimit {
	mgr.lock.RLock()
	rLimit, ok := mgr.maps[key]
	if !ok {
		mgr.lock.RUnlock()
		rLimit = newRateLimit(limit, rate)

		// save limit
		mgr.lock.Lock()
		mgr.maps[key] = rLimit
		mgr.lock.Unlock()
	} else {
		mgr.lock.RUnlock()
	}

	return rLimit
}

type rateLimit struct {
	Remaining int32         // 剩余多少次
	Reset     int64         // 下次重置时间
	limit     int32         // 重置限制次数
	rate      time.Duration // 限制区间
	lock      sync.Mutex
}

func newRateLimit(limit int, rate time.Duration) *rateLimit {
	rLimit := &rateLimit{
		Remaining: int32(limit),
		Reset:     time.Now().Add(rate).Unix(),
		limit:     int32(limit),
		rate:      rate,
		lock:      sync.Mutex{},
	}
	return rLimit
}

func (l *rateLimit) Get() bool {
	l.lock.Lock()
	now := time.Now()
	if now.Unix() > l.Reset {
		l.Remaining = l.limit
		l.Reset = now.Add(l.rate).Unix()
	}
	l.Remaining -= 1
	l.lock.Unlock()
	return l.Remaining >= 0
}
