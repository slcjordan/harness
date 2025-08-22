//go:build workcmd

package main

import (
	"context"

	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var WorkInit []func(context.Context, worker.Worker) error
var WorkOptions []cli.Option

func init() {
	config.Workflow.QueueName = "harness-worker"

	cmd.Subcommand(
		"activity", "run temporal activity worker", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
			workerClient, err := client.Dial(client.Options{
				HostPort:  config.Workflow.Server,
				Namespace: "default",
			})
			if err != nil {
				return err
			}
			defer workerClient.Close()
			w := worker.New(workerClient, config.Workflow.QueueName, worker.Options{})
			for _, f := range WorkInit {
				err = f(ctx, w)
				if err != nil {
					return err
				}
			}
			return w.Run(worker.InterruptCh())
		}), append(WorkOptions, cli.WithWorkflowFlags)...)
}
