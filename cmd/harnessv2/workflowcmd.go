//go:build workflowcmd

package main

import (
	"context"

	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var WorkflowInit []func(worker.Worker) error

func init() {
	cmd.Subcommand(
		"run-workflow-worker", "run temporal workflow worker", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
			workerClient, err := client.Dial(client.Options{
				HostPort:  config.Workflow.Server,
				Namespace: "default",
			})
			if err != nil {
				return err
			}
			defer workerClient.Close()
			w := worker.New(workerClient, config.Workflow.QueueName, worker.Options{})
			for _, f := range WorkflowInit {
				err = f(w)
				if err != nil {
					return err
				}
			}
			return nil
		}), WorkOptions...)
}
