package main

import (
	"context"
	"os"

	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/logger"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	logger.Init()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := cli.NewCommand("run", "run background worker", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
		workerClient, err := client.Dial(client.Options{})
		if err != nil {
			return err
		}
		defer workerClient.Close()

		w := worker.New(workerClient, config.Workflow.QueueName, worker.Options{})
		return w.Run(worker.InterruptCh())
	}))

	err := c.Run(ctx, os.Args)
	if err != nil {
		logger.Infof(ctx, "application error: %s", err)
	}
}
