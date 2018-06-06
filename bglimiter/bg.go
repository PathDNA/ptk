package bglimiter

import (
	"context"
	"sync"
	"time"
)

// New returns a new runner with no limits.
func New() *BackgroundLimiter { return NewWithContext(context.Background(), 0) }

// NewWithContext returns a new runner with the given parent context and limit, if limit is <= 0 it won't have a limit.
func NewWithContext(ctx context.Context, limit int) *BackgroundLimiter {
	var bg BackgroundLimiter
	bg.ctx, bg.cancel = context.WithCancel(ctx)

	if limit > 0 {
		bg.ch = make(chan struct{}, limit)
	}

	return &bg
}

type BackgroundLimiter struct {
	wg sync.WaitGroup
	ch chan struct{}

	ctx    context.Context
	cancel func()
}

func (bg *BackgroundLimiter) Add(fn func(ctx context.Context) error) <-chan error {
	errChan := make(chan error, 2)
	if !bg.add() {
		errChan <- context.Canceled
		close(errChan)
		return errChan
	}

	go func() { errChan <- fn(bg.ctx) }()

	go func() {
		select {
		case err := <-errChan: // instead of an extra communication channel, don't judge
			errChan <- err
		case <-bg.ctx.Done():
			errChan <- bg.ctx.Err()
		}
		bg.done()
	}()

	return errChan
}

func (bg *BackgroundLimiter) AddWithTimeout(fn func(ctx context.Context) error, timeout time.Duration) <-chan error {
	errChan := make(chan error, 2)

	if !bg.add() {
		errChan <- context.Canceled
		close(errChan)
		return errChan
	}

	tctx, cancel := context.WithTimeout(bg.ctx, timeout)

	go func() { errChan <- fn(tctx) }()

	go func() {
		select {
		case err := <-errChan: // instead of an extra communication channel, don't judge
			errChan <- err
		case <-bg.ctx.Done():
			errChan <- bg.ctx.Err()
		case <-tctx.Done():
			errChan <- tctx.Err()
		}
		cancel()
		bg.done()
	}()

	return errChan
}

func (bg *BackgroundLimiter) Context() context.Context { return bg.ctx }

func (bg *BackgroundLimiter) Close() error {
	err := bg.ctx.Err()
	bg.cancel()
	return err
}

func (bg *BackgroundLimiter) IsCanceled() bool {
	return bg.ctx.Err() != nil
}

func (bg *BackgroundLimiter) Wait() { bg.wg.Wait() }
func (bg *BackgroundLimiter) WaitWithContext(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		bg.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-bg.ctx.Done():
		return bg.ctx.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (bg *BackgroundLimiter) add() bool {
	if bg.ch != nil {
		bg.ch <- struct{}{}
	}

	if bg.IsCanceled() {
		return false
	}

	bg.wg.Add(1)
	return true
}

func (bg *BackgroundLimiter) done() {
	if bg.ch != nil {
		<-bg.ch
	}

	bg.wg.Done()
}
