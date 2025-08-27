package scheduler

import (
	"context"
	"time"

	"github.com/slcjordan/harness"
)

type Priority struct {
	high    chan func()
	low     chan func()
	size    int
	weight  float64 // [0-1)
	timeout time.Duration
}

// weight determines what percentge of threads should give preference to low-priority tasks [0-1).
func NewPriority(size int, weight float64, timeout time.Duration) *Priority {
	return &Priority{
		high:    make(chan func(), size),
		low:     make(chan func(), 1),
		size:    size,
		weight:  weight,
		timeout: timeout,
	}
}

func schedule(ctx context.Context, ch chan func(), f func(), timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return harness.Errorf(harness.ErrContextCancelled, "context cancelled")
	case <-timer.C:
		return harness.Errorf(harness.ErrTimeout, "timed out")
	case ch <- f:
		return nil
	}
}

func (p *Priority) ScheduleHigh(ctx context.Context, f func()) error {
	return schedule(ctx, p.high, f, p.timeout)
}

func (p *Priority) ScheduleLow(ctx context.Context, f func()) error {
	return schedule(ctx, p.low, f, p.timeout)
}

func (p *Priority) Run(ctx context.Context) {
	for i := 0; i < p.size; i++ {
		if float64(i) < p.weight*float64(p.size) {
			go (&thread{
				first:  p.low,
				second: p.high,
			}).run(ctx)
		} else {
			go (&thread{
				first:  p.high,
				second: p.low,
			}).run(ctx)
		}
	}
}

type thread struct {
	first  chan func()
	second chan func()
}

func (t *thread) run(ctx context.Context) {
	for {

		// check if done
		select {
		case <-ctx.Done():
			return
		default:
		}

		// prioritize first
		select {
		case f := <-t.first:
			f()
			continue
		default:
		}

		// take whichever has work available
		select {
		case <-ctx.Done():
			return
		case f := <-t.first:
			f()
		case f := <-t.second:
			f()
		}
	}
}
