package exec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
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
	ctx = logger.With(ctx, "path", path)
	ctx = logger.With(ctx, "args", args)
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
		go result.notify(harness.Stdout, io.TeeReader(stdout, os.Stdout))
		go result.notify(harness.Stderr, io.TeeReader(stderr, os.Stderr))
		err = cmd.Start()
		go io.WriteString(stdin, "k")
		if err != nil {
			logger.Errorf(ctx, "%q could not start: %s", cmd, err)
			return
		}

		go func() {
			for {
				select {
				case data := <-result.input:
					input := make([]byte, len(data))
					copy(input, data)
					fmt.Printf("%q\n", string(input))
					io.Copy(stdin, bytes.NewReader(input))
				case <-ctx.Done():
					return
				}
			}
		}()
		err = cmd.Wait()
		if err != nil {
			logger.Errorf(ctx, "%q while running: %s", cmd, err)
			return
		}
		logger.Infof(ctx, "application exited")
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
	var buff [1024]byte
	for {
		n, err := r.Read(buff[:])
		if err != nil {
			logger.Errorf(i.ctx, "error while scanning stream %d: %s", s, err)
			return
		}
		line := make([]byte, n)
		copy(line, buff[:])
		logger.Infof(i.ctx, "[%d] %s", s, string(line))
		go i.notifyListeners(harness.CommandEvent{Stream: s, Data: line})
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
