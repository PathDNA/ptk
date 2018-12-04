package ptk

import (
	"context"
	"time"
)

func WaitFor(fn func() error) <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- fn()
		close(ch)
	}()
	return ch
}

func SleepUntil(ctx context.Context, hour, min, sec int) error {
	var (
		now    = time.Now().UTC()
		target = time.Date(now.Year(), now.Month(), now.Day(), hour, min, sec, 0, time.UTC)
	)

	if target.Before(now) {
		target = target.AddDate(0, 0, 1)
	}

	if ctx == nil {
		ctx = context.Background()
	}

	t := time.NewTimer(target.Sub(now))

	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		if !t.Stop() {
			<-t.C
		}
		return ctx.Err()
	}
}
