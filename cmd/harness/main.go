package main

import (
	"context"
	"os"

	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/logger"
)

func init() {
	config.Workflow.Server = "workflow-server:7233"
}

var WorkOptions []cli.Option
var ServeOptions []cli.Option

var (
	GitlabListMRs = harness.Contract[string, harness.MergeRequest]{Queue: "gitlab", Name: "list-mrs"}
	UserMessages  = harness.Contract[[]harness.UserMessage, struct{}]{Queue: "gitlab", Name: "user-messages"}
)

var cmd = cli.NewCommand("harness", "do nothing", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
	return nil
}))

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cmd.Run(ctx, os.Args)
	if err != nil {
		logger.Infof(ctx, "application error: %s", err)
	}
}
