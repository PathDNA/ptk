package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"sync"
	"time"
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
	m    sync.RWMutex
}

func (mc *MemCache) Set(key string, val interface{}, ttl time.Duration) error {
	mc.m.Lock()
	defer mc.m.Unlock()

	if mc.c == nil {
		return os.ErrClosed
	}
	mc.c[key] = &cacheItem{val, time.Now().Add(ttl).Unix()}

	return nil
}

func (mc *MemCache) Delete(key string) error {
	mc.m.Lock()
	defer mc.m.Unlock()

	if mc.c == nil {
		return os.ErrClosed
	}
	delete(mc.c, key)

	return nil
}

func (mc *MemCache) Get(key string) (val interface{}, found bool) {
	var ci *cacheItem
	mc.m.RLock()
	if ci, found = mc.c[key]; found {
		val = ci.Value
	}
	mc.m.RUnlock()

	return
}

func (mc *MemCache) GetAndDelete(key string) (val interface{}, found bool) {
	var ci *cacheItem
	mc.m.Lock()
	if ci, found = mc.c[key]; found {
		val = ci.Value
		delete(mc.c, key)
	}
	mc.m.Unlock()

	return
}

func (mc *MemCache) Clean() (n int) {
	now := time.Now().Unix()
	mc.m.Lock()
	for k, ci := range mc.c {
		if now > ci.ExpiresAt {
			delete(mc.c, k)
			n++
		}
	}
	mc.m.Unlock()

	return
}

func (mc *MemCache) Reset() {
	mc.m.Lock()
	mc.c = map[string]*cacheItem{}
	mc.m.Unlock()
}

func (mc *MemCache) Close() error {
	select {
	case <-mc.done:
		return os.ErrClosed
	default:
	}

	mc.m.Lock()
	close(mc.done)
	mc.c = nil
	mc.m.Unlock()

	return nil
}

func Key(args ...string) string {
	h := sha1.New()
	for _, a := range args {
		io.WriteString(h, a)
	}

	return hex.EncodeToString(h.Sum(nil))
}

type cacheItem struct {
	Value     interface{}
	ExpiresAt int64
}
