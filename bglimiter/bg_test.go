package bglimiter_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/PathDNA/ptk/bglimiter"
)

func TestLimit(t *testing.T) {
	var (
		n  int64
		bg = bglimiter.NewWithContext(context.Background(), 2)
	)

	fn := func(ctx context.Context) error {
		atomic.AddInt64(&n, 1)
		<-ctx.Done()
		return nil
	}

	go func() {
		bg.Add(fn)
		bg.Add(fn)

		// fn will no longer run since when .Add executes, bg is canceled
		bg.Add(fn)
		bg.Add(fn)
		atomic.StoreInt64(&n, 99)
	}()

	time.Sleep(time.Millisecond * 2)
	if nn := atomic.LoadInt64(&n); nn != 2 {
		t.Fatalf("expected 2, got %d", nn)
	}

	bg.Close()
	time.Sleep(time.Millisecond)

	if nn := atomic.LoadInt64(&n); nn != 99 {
		t.Fatalf("expected 99, got %d", nn)
	}

	bg = bglimiter.NewWithContext(context.Background(), 1)
	fn = func(ctx context.Context) error {
		atomic.AddInt64(&n, 1)
		time.Sleep(5 * time.Millisecond)
		return nil
	}

	errCh := bg.AddWithTimeout(fn, time.Millisecond)
	select {
	case err := <-errCh:
		if err != context.DeadlineExceeded {
			t.Fatalf("expected Deadline, got %v", err)
		}
	case <-time.After(time.Millisecond * 5):
		t.Fatal("didn't timeout :(")
	}

	if bg.IsCanceled() {
		t.Fatal("bg.IsCanceled() :( :(")
	}
}
