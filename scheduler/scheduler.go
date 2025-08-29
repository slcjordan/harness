package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/slcjordan/harness"
)

type Job struct {
	Do     func()
	Cancel func()
}

type Priority struct {
	mu      sync.RWMutex // guards channels so schedulers don't publish to a closing channel.
	high    chan Job
	low     chan Job
	size    int
	weight  float64 // [0-1)
	timeout time.Duration
}

// weight determines what percentage of threads should give preference to low-priority tasks [0-1).
func NewPriority(size int, weight float64, timeout time.Duration) *Priority {
	return &Priority{
		high:    make(chan Job, size),
		low:     make(chan Job, 1),
		size:    size,
		weight:  weight,
		timeout: timeout,
	}
}

func schedule(ch chan Job, f func(), cancel func(), timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		cancel()
	case ch <- Job{Do: f, Cancel: cancel}:
	}
}

func (p *Priority) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	oldHigh, oldLow := p.high, p.low
	defer close(oldHigh)
	defer close(oldLow)

	p.low = nil
	p.high = nil

	timer := time.NewTimer(p.timeout)
	defer timer.Stop()

	for {
		select {
		case curr := <-oldLow:
			go curr.Cancel()
		case curr := <-oldHigh:
			go curr.Cancel()
		case <-timer.C:
			return
		}
	}
}

func (p *Priority) ScheduleHigh(f func(), cancel func()) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	schedule(p.high, f, cancel, p.timeout)
}

func (p *Priority) ScheduleLow(f func(), cancel func()) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	schedule(p.low, f, cancel, p.timeout)
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
	first  chan Job
	second chan Job
}

func (t *thread) run(ctx context.Context) {
	first, second := t.first, t.second
	for {

		// check if done
		select {
		case <-ctx.Done():
			return
		default:
		}

		// prioritize first
		select {
		case curr, ok := <-first:
			if ok {
				curr.Do()
			} else {
				first = nil
			}
			continue
		default:
		}

		// take whichever has work available
		select {
		case <-ctx.Done():
			return
		case curr, ok := <-first:
			if ok {
				curr.Do()
			} else {
				first = nil
			}
		case curr, ok := <-second:
			if ok {
				curr.Do()
			} else {
				second = nil
			}
		}
	}
}

type HandlerFunc[Ctx, Input, Output any] func(Ctx, Input) (Output, error)

func (h HandlerFunc[Ctx, Input, Output]) Handle(ctx Ctx, input Input) (Output, error) {
	return h(ctx, input)
}

func Wrap[Ctx, Input, Output any](
	c harness.Contract[Input, Output],
	handler HandlerFunc[Ctx, Input, Output],
	p *Priority,
) HandlerFunc[Ctx, Input, Output] {

	return func(ctx Ctx, input Input) (out Output, err error) {
		done := make(chan struct{})
		f := func() {
			defer close(done)
			out, err = handler.Handle(ctx, input)
		}
		cancel := func() {
			defer close(done)
			err = harness.Errorf(harness.ErrTimeout, "scheduler timeout")
		}
		switch c.Priority {
		case harness.High:
			p.ScheduleHigh(f, cancel)
		case harness.Low:
			p.ScheduleLow(f, cancel)
		}
		<-done
		return
	}
}
