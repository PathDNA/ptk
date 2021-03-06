package cache

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PathDNA/atoms"
)

func NewMemCache(autoCleanEvery time.Duration) (mc *MemCache) {
	mc = &MemCache{
		c:    map[string]*cacheItem{},
		done: make(chan struct{}, 1),
	}

	if autoCleanEvery > 0 {
		t := time.NewTicker(autoCleanEvery)
		go func() {
			select {
			case <-mc.done:
				t.Stop()
				return
			case <-t.C:
				mc.Clean()
			}

		}()
	}

	return
}

type MemCache struct {
	c    map[string]*cacheItem
	done chan struct{}
	// using double mutexes to handle long updates
	mux  sync.RWMutex
	mmux atoms.MultiMux
}

func (mc *MemCache) Set(key string, val interface{}, ttl time.Duration) (err error) {
	mc.mmux.Update(key, func() {
		mc.mux.Lock()
		if mc.c == nil {
			err = os.ErrClosed
		} else {
			mc.c[key] = &cacheItem{Value: val, ExpiresAt: time.Now().Add(ttl).Unix()}
		}
		mc.mux.Unlock()
	})

	return
}

// if ttl is -1, the key gets deleted
// if keepOldTTL is true, the original expiry ts will be kept
func (mc *MemCache) Update(key string, fn func(old interface{}) (val interface{}, keepOldTTL bool, ttl time.Duration)) (err error) {
	var del bool
	mc.mmux.Update(key, func() {
		var old interface{}

		mc.mux.RLock()
		if ci := mc.c[key]; ci != nil {
			old = ci.Value
		}
		mc.mux.RUnlock()

		val, keepOldTTL, ttl := fn(old)

		mc.mux.Lock()
		defer mc.mux.Unlock()
		if del = ttl == -1; del {
			delete(mc.c, key)
		}
		if mc.c == nil {
			err = os.ErrClosed
		} else {
			if keepOldTTL {
				if ci := mc.c[key]; ci != nil {
					ci.Value = val
					return
				}
			}
			mc.c[key] = &cacheItem{Value: val, ExpiresAt: time.Now().Add(ttl).Unix()}
		}
	})

	if del {
		mc.mmux.Delete(key)
	}

	return
}

func (mc *MemCache) Delete(key string) (err error) {
	mc.mmux.Update(key, func() {
		mc.mux.Lock()
		if mc.c == nil {
			err = os.ErrClosed
		} else {
			delete(mc.c, key)
		}
		mc.mux.Unlock()
	})
	mc.mmux.Delete(key)
	return
}

func (mc *MemCache) Get(key string) (val interface{}, found bool) {
	mc.mmux.Read(key, func() {
		var ci *cacheItem
		mc.mux.RLock()
		if ci, found = mc.c[key]; found {
			val = ci.Value
		}
		mc.mux.RUnlock()
	})

	return
}

func (mc *MemCache) Clean() (n int) {
	now := time.Now().Unix()
	mc.mux.Lock()
	for k, ci := range mc.c {
		if now > ci.ExpiresAt {
			delete(mc.c, k)
			go mc.mmux.Delete(k)
			n++
		}
	}
	mc.mux.Unlock()

	return
}

func (mc *MemCache) Reset() {
	mc.mux.Lock()
	mc.c = map[string]*cacheItem{}
	mc.mux.Unlock()
}

func (mc *MemCache) Close() error {
	select {
	case <-mc.done:
		return os.ErrClosed
	default:
	}

	mc.mux.Lock()
	close(mc.done)
	mc.c = nil
	mc.mux.Unlock()

	return nil
}

func Key(args ...interface{}) string {
	if len(args) == 0 {
		panic("this shouldn't happen")
	}

	var b strings.Builder
	for _, a := range args {
		fmt.Fprintf(&b, "%v:", a)
	}

	return b.String()[:b.Len()-1]
}

type cacheItem struct {
	Value     interface{}
	ExpiresAt int64
}
