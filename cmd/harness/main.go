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

var (
	GitlabListMRs             = harness.Contract[string, []harness.MergeRequest]{Queue: "gitlab", Name: "list-mrs"}
	GitlabUserMessages        = harness.Contract[[]harness.MergeRequest, []harness.UserMessage]{Queue: "gitlab", Name: "build-mr-user-messages"}
	SlackUserMessages         = harness.Contract[[]harness.UserMessage, struct{}]{Queue: "slack", Name: "send-user-messages"}
	WorkflowGitlabNotifySlack = harness.Contract[string, struct{}]{Queue: "coordinator", Name: "fetch-mrs-and-send-user-messages"}
)

var cmd = cli.NewCommand("harness", "do nothing", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
	return nil
}))

func main() {
	logger.Init()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cmd.Run(ctx, os.Args)
	if err != nil {
		logger.Infof(ctx, "application error: %s", err)
	}
}
