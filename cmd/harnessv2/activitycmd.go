//go:build activitycmd

package main

import (
	"context"

	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var ActivityInit []func(worker.Worker) error

func init() {
	config.Workflow.QueueName = "harness-worker"

	cmd.Subcommand(
		"run-activity-worker", "run temporal activity worker", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
			workerClient, err := client.Dial(client.Options{
				HostPort:  config.Workflow.Server,
				Namespace: "default",
			})
			if err != nil {
				return err
			}
			defer workerClient.Close()
			w := worker.New(workerClient, config.Workflow.QueueName, worker.Options{})
			for _, f := range ActivityInit {
				err = f(w)
				if err != nil {
					return err
				}
			}
			return nil
		}), append(WorkOptions, cli.WithWorkflowFlags)...)
}
