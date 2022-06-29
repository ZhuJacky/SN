// Package limiter provides ...
package limiter

import (
	"sync"
	"time"
)

// LimitMap store the rateLimit
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

// Load a rateLimit, if not exist will create
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
	Remaining int32         // remain count
	Reset     int64         // next reset limit, seconds
	limit     int32         // limit count
	rate      time.Duration // time expire
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

// Get the permission to access the api
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
