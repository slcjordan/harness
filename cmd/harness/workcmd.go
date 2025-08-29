//go:build workcmd

package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/scheduler"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var WorkInit []func(context.Context, func(queueName string) worker.Worker, *scheduler.Priority) error
var WorkOptions []cli.Option

func RegisterWorkflowHook(f func(context.Context, func(queueName string) worker.Worker, *scheduler.Priority) error) {
	WorkInit = append(WorkInit, f)
}

func RegisterOption(options ...cli.Option) {
	WorkOptions = append(WorkOptions, options...)
}

func init() {
	config.Workflow.QueueName = "harness-worker"

	cmd.Subcommand(
		"work", "run temporal activity worker", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
			workerClient, err := client.Dial(client.Options{
				HostPort:  config.Workflow.Server,
				Namespace: "default",
			})
			if err != nil {
				return err
			}
			defer workerClient.Close()
			cache := make(map[string]worker.Worker)
			p := scheduler.NewPriority(64, 0.2, time.Second)
			for _, f := range WorkInit {
				err = f(ctx, func(queueName string) worker.Worker {
					w, ok := cache[queueName]
					if !ok {
						w = worker.New(workerClient, queueName, worker.Options{})
						cache[queueName] = w
					}
					return w
				}, p)
				if err != nil {
					return err
				}
			}

			var wg sync.WaitGroup
			var mu sync.Mutex
			var errs []error
			for _, w := range cache {
				wg.Add(1)

				go func() {
					defer wg.Done()

					err := w.Run(worker.InterruptCh())

					mu.Lock()
					defer mu.Unlock()
					if err != nil {
						errs = append(errs, err)
					}
				}()
			}
			wg.Wait()
			if errs != nil {
				return errors.Join(errs...)
			}
			return nil
		}), append(WorkOptions, cli.WithWorkflowFlags)...)
}
