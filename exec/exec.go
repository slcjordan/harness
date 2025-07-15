package exec

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/logger"
)

const timeout = 5 * time.Second

type Interactive struct {
	ctx   context.Context
	cmd   *exec.Cmd
	input chan []byte

	mu          sync.RWMutex
	subscribers map[string]struct{}
	listener    harness.Notifier[harness.CommandEvent]
}

func StartInteractive(ctx context.Context, listener harness.Notifier[harness.CommandEvent], path string, args ...string) *Interactive {
	result := &Interactive{
		ctx:         ctx,
		input:       make(chan []byte),
		subscribers: make(map[string]struct{}),
		listener:    listener,
	}

	go func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		cmd := exec.CommandContext(ctx, path, args...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			logger.Errorf(ctx, "%q could not pipe stdin: %s", cmd, err)
			return
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			logger.Errorf(ctx, "%q could not pipe stdout: %s", cmd, err)
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			logger.Errorf(ctx, "%q could not pipe stderr: %s", cmd, err)
			return
		}
		err = cmd.Start()
		if err != nil {
			logger.Errorf(ctx, "%q could not start: %s", cmd, err)
			return
		}
		go result.notify(harness.Stdout, stdout)
		go result.notify(harness.Stderr, stderr)

		go func() {
			for {
				select {
				case data := <-result.input:
					io.Copy(stdin, bytes.NewReader(data))
				case <-ctx.Done():
					return
				}
			}
		}()
		cmd.Wait()
	}()

	return result
}

func (i *Interactive) Done() <-chan struct{} {
	return i.ctx.Done()
}

func (i *Interactive) notifyListeners(evt harness.CommandEvent) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	for id := range i.subscribers {
		go i.listener.Notify(id, evt)
	}
}

func (i *Interactive) notify(s harness.Stream, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		go i.notifyListeners(harness.CommandEvent{Stream: s, Data: line})
	}
	if err := scanner.Err(); err != nil {
		logger.Errorf(context.TODO(), "%q: error while scanning stream %d: %s", i.cmd.String(), s, err)
	}

}

func (i *Interactive) Attach(id string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.subscribers[id] = struct{}{}
}

func (i *Interactive) Unattach(id string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	delete(i.subscribers, id)
}

func (i *Interactive) Handle(ctx context.Context, input harness.InteractiveInput) (struct{}, error) {
	select {
	case i.input <- input.Data:
	case <-ctx.Done():
		return struct{}{}, harness.Errorf(harness.ErrContextCancelled, "could not process interactive input: %w", ctx.Err())
	case <-time.After(timeout):
		return struct{}{}, harness.Errorf(harness.ErrTimeout, "could not process interactive input: timed out after %s", timeout)
	}
	return struct{}{}, nil
}
