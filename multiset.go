package ptk

import "sync"

// Set is a simple set.
type Set map[string]struct{}

func (s *Set) Set(key string) {
	if *s == nil {
		*s = make(map[string]struct{})
	}
	(*s)[key] = struct{}{}
}

func (s Set) Delete(key string) {
	delete(s, key)
}

func (s Set) Has(key string) bool {
	_, ok := s[key]
	return ok
}

func (s Set) Keys() []string {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}

// MultiSet is a very simple Set
type MultiSet map[string]Set

func (ms *MultiSet) Set(key, subKey string) {
	s := ms.Sub(key)
	s.Set(subKey)
}

func (ms MultiSet) Delete(key, subKey string) {
	ms[key].Delete(subKey)
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

func (ms *MultiSet) Sub(key string) Set {
	m := (*ms)[key]
	if m == nil {
		if *ms == nil {
			*ms = make(map[string]Set)
		}
		m = Set{}
		(*ms)[key] = m
	}
	return m
}

// SafeMultiSet is a concurrent-safe MultiSet
type SafeMultiSet struct {
	ms  MultiSet
	mux sync.RWMutex
}

func (sms *SafeMultiSet) Set(key, subKey string) {
	sms.mux.Lock()
	sms.ms.Set(key, subKey)
	sms.mux.Unlock()
}

func (sms *SafeMultiSet) Delete(key, subKey string) {
	sms.mux.Lock()
	sms.ms.Delete(key, subKey)
	sms.mux.Unlock()
}

func (sms *SafeMultiSet) Has(key, subKey string) bool {
	sms.mux.RLock()
	ok := sms.ms.Has(key, subKey)
	sms.mux.RUnlock()
	return ok
}

func (sms *SafeMultiSet) Keys() []string {
	sms.mux.RLock()
	keys := sms.ms.Keys()
	sms.mux.RUnlock()
	return keys
}

func (sms *SafeMultiSet) Values(key string) []string {
	sms.mux.RLock()
	vals := sms.ms.Values(key)
	sms.mux.RUnlock()
	return vals
}
