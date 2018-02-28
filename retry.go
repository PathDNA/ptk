package ptk

import (
	"context"
	"time"
)

// Retry is an alias for RetryWithCtx(context.Background(), fn, attempts, delay, backoffMod)
func Retry(fn func() error, attempts uint, delay time.Duration, backoffMod uint) error {
	return RetryWithCtx(context.Background(), fn, attempts, delay, backoffMod)
}

// RetryWithCtx calls fn every (delay * backoffMod) until it returns nil, the passed ctx is done or attempts are reached.
func RetryWithCtx(ctx context.Context, fn func() error, attempts uint, delay time.Duration, backoffMod uint) error {
	if delay == 0 {
		delay = time.Second
	}

	if attempts == 0 {
		attempts = 1
	}

	if backoffMod == 0 {
		backoffMod = 1
	}

	ret := make(chan error, 1)

	go func() {
		var err error
		for ; attempts > 0; attempts-- {
			if err = fn(); err == nil {
				break
			}
			time.Sleep(delay)
			delay = delay * time.Duration(backoffMod)
		}
		ret <- err
	}()

	select {
	case err := <-ret:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}