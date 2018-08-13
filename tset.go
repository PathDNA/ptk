package ptk

import (
	"os"
	"sync"
	"time"
)

func NewTimeoutSet(purgeTimeout time.Duration) *TimeoutSet {
	ss := &TimeoutSet{
		s:    map[string]int64{},
		done: make(chan struct{}, 1),
	}

	go func() {
		for {
			select {
			case <-ss.done:
				return
			case <-time.After(purgeTimeout):
			}
			ss.mux.Lock()
			now := time.Now().UnixNano()
			for k, t := range ss.s {
				if t > -1 && t <= now {
					delete(ss.s, k)
				}
			}
			ss.mux.Unlock()
		}
	}()

	return ss
}

type TimeoutSet struct {
	s    map[string]int64
	mux  sync.RWMutex
	done chan struct{}
}

func (ss *TimeoutSet) Set(key string, to time.Duration) {
	ts := time.Now().Add(to).UnixNano()
	ss.mux.Lock()
	ss.s[key] = ts
	ss.mux.Unlock()
}

func (ss *TimeoutSet) SetAny(keys ...string) {
	ss.mux.Lock()
	for _, k := range keys {
		ss.s[k] = -1
	}
	ss.mux.Unlock()
}

func (ss *TimeoutSet) Delete(keys ...string) {
	ss.mux.Lock()
	for _, k := range keys {
		delete(ss.s, k)
	}
	ss.mux.Unlock()
}

func (ss *TimeoutSet) Has(key string) bool {
	now := time.Now().UnixNano()
	ss.mux.RLock()
	t := ss.s[key]
	ss.mux.RUnlock()
	return t == -1 || t > now
}

func (ss *TimeoutSet) Len() int {
	ss.mux.RLock()
	ln := len(ss.s)
	ss.mux.RUnlock()
	return ln
}

func (ss *TimeoutSet) Keys() []string {
	now := time.Now().UnixNano()
	ss.mux.RLock()
	keys := make([]string, 0, len(ss.s))
	for k, t := range ss.s {
		if t == -1 || t > now {
			keys = append(keys, k)
		}
	}
	ss.mux.RUnlock()
	return keys
}

func (ss *TimeoutSet) IsClosed() bool {
	select {
	case <-ss.done:
		return true
	default:
		return false
	}
}

func (ss *TimeoutSet) Close() error {
	ss.mux.Lock()
	defer ss.mux.Unlock()

	select {
	case <-ss.done:
		return os.ErrClosed
	default:
		close(ss.done)
		return nil
	}
}
