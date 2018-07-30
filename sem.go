package ptk

import (
	"fmt"
	"sync"
)

func NewSem(size int) *Sem {
	if size == 0 {
		size = 1
	}

	return &Sem{ch: make(chan struct{}, size)}
}

type Sem struct {
	wg sync.WaitGroup
	ch chan struct{}
}

func (s *Sem) Add(n int) {
	if n == 0 {
		return
	}

	if n > 0 {
		var e struct{}
		for i := 0; i < n; i++ {
			s.ch <- e
		}
	} else {
		for i := 0; i > n; i-- {
			<-s.ch
		}
	}

	s.wg.Add(n)
}

func Add(n int) {
	if n == 0 {
		return
	}

	if n > 0 {
		for i := 0; i < n; i++ {
			fmt.Println(i, n, i-n, n-i)
		}
	} else {
		for i := n; i > 0; i-- {
			fmt.Println(i, n, i-n, n-i)
		}
	}
}

func (s *Sem) Done() {
	s.Add(-1)
}

func (s *Sem) Wait() {
	s.wg.Wait()
}

func (s *Sem) Close() { close(s.ch) }
