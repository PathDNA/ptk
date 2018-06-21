package ptk

import "sync"

// Set is a simple set.
type Set map[string]struct{}

func (s Set) Set(keys ...string) {
	var e struct{}
	for _, k := range keys {
		s[k] = e
	}
}

func (s Set) Delete(keys ...string) {
	for _, k := range keys {
		delete(s, k)
	}
}

func (s Set) Has(key string) bool {
	_, ok := s[key]
	return ok
}

func (s Set) Len() int {
	return len(s)
}

func (s Set) Keys() []string {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}

func NewSafeSet(keys ...string) *SafeSet {
	s := Set{}
	s.Set(keys...)

	return &SafeSet{
		s: s,
	}
}

type SafeSet struct {
	s   Set
	mux sync.RWMutex
}

func (ss *SafeSet) Set(keys ...string) {
	ss.mux.Lock()
	ss.s.Set(keys...)
	ss.mux.Unlock()
}

func (ss *SafeSet) Delete(keys ...string) {
	ss.mux.Lock()
	ss.s.Delete(keys...)
	ss.mux.Unlock()
}

func (ss *SafeSet) Has(key string) bool {
	ss.mux.RLock()
	ok := ss.s.Has(key)
	ss.mux.RUnlock()
	return ok
}

func (ss *SafeSet) Len() int {
	ss.mux.RLock()
	ln := ss.s.Len()
	ss.mux.RUnlock()
	return ln
}

func (ss *SafeSet) Keys() []string {
	ss.mux.RLock()
	keys := ss.s.Keys()
	ss.mux.RUnlock()
	return keys
}
