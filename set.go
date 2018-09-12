package ptk

import (
	"encoding/json"
	"sort"
	"sync"
)

func NewSet(keys ...string) Set {
	s := Set{}
	s.Set(keys...)
	return s
}

// Set is a simple set.
type Set map[string]struct{}

func (s *Set) Set(keys ...string) {
	sm := *s
	if sm == nil {
		sm = Set{}
		*s = sm
	}

	var e struct{}
	for _, k := range keys {
		sm[k] = e
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

func (s Set) MarshalJSON() ([]byte, error) {
	keys := s.Keys()
	sort.Strings(keys)
	return json.Marshal(keys)
}

func (s *Set) UnmarshalJSON(data []byte) (err error) {
	var keys []string
	if err = json.Unmarshal(data, &keys); err == nil {
		s.Set(keys...)
	}
	return
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

func (ss *SafeSet) MarshalJSON() ([]byte, error) {
	ss.mux.RLock()
	b, err := ss.s.MarshalJSON()
	ss.mux.RUnlock()
	return b, err
}

func (ss *SafeSet) UnmarshalJSON(data []byte) error {
	ss.mux.Lock()
	err := ss.s.UnmarshalJSON(data)
	ss.mux.Unlock()
	return err
}
