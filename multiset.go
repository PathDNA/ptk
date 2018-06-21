package ptk

import "sync"

func NewMultiSet(keys ...string) MultiSet {
	ms := MultiSet{}
	for _, key := range keys {
		ms[key] = Set{}
	}
	return ms
}

// MultiSet is a very simple Set
type MultiSet map[string]Set

func (ms MultiSet) Set(key string, subKeys ...string) {
	ms.Sub(key).Set(subKeys...)
}

func (ms MultiSet) Delete(key string, subKeys ...string) {
	ms[key].Delete(subKeys...)
}

func (ms MultiSet) Has(key, subKey string) bool {
	return ms[key].Has(subKey)
}

func (ms MultiSet) Keys() []string {
	keys := make([]string, 0, len(ms))
	for k := range ms {
		keys = append(keys, k)
	}
	return keys
}

func (ms MultiSet) Values(key string) []string {
	return ms[key].Keys()
}

func (ms MultiSet) Sub(key string) Set {
	m := ms[key]
	if m == nil {
		m = Set{}
		ms[key] = m
	}
	return m
}

func NewSafeMultiSet(keys ...string) *SafeMultiSet {
	ms := map[string]*SafeSet{}
	for _, key := range keys {
		ms[key] = NewSafeSet()
	}
	return &SafeMultiSet{
		ms: ms,
	}
}

// SafeMultiSet is a concurrent-safe MultiSet
type SafeMultiSet struct {
	ms  map[string]*SafeSet
	mux sync.RWMutex
}

func (sms *SafeMultiSet) Set(key string, subKeys ...string) {
	sms.Sub(key).Set(subKeys...)
}

func (sms *SafeMultiSet) Delete(key string, subKeys ...string) {
	if ss := sms.rsub(key); ss != nil {
		ss.Delete(subKeys...)
	}
}

func (sms *SafeMultiSet) Has(key, subKey string) (ok bool) {
	if ss := sms.rsub(key); ss != nil {
		ok = ss.Has(subKey)
	}
	return
}

func (sms *SafeMultiSet) Keys() (keys []string) {
	sms.mux.RLock()
	keys = make([]string, 0, len(sms.ms))
	for k := range sms.ms {
		keys = append(keys, k)
	}
	sms.mux.RUnlock()
	return
}

func (sms *SafeMultiSet) Values(key string) (vals []string) {
	if ss := sms.rsub(key); ss != nil {
		vals = ss.Keys()
	}
	return
}

func (sms *SafeMultiSet) Sub(key string) (ss *SafeSet) {
	if ss = sms.rsub(key); ss != nil {
		return
	}

	sms.mux.Lock()
	if ss = sms.ms[key]; ss == nil {
		ss = NewSafeSet()
		sms.ms[key] = ss
	}
	sms.mux.Unlock()

	return ss
}

func (sms *SafeMultiSet) rsub(key string) (ss *SafeSet) {
	sms.mux.RLock()
	ss = sms.ms[key]
	sms.mux.RUnlock()
	return
}
