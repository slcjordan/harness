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
	GitlabListMRs = harness.Contract[
		string,
		[]harness.MergeRequest,
	]{
		Domain:   "gitlab",
		Name:     "list-mrs",
		Version:  harness.V1,
		Tier:     harness.Admin,
		Priority: harness.Low,
		Strategy: harness.Sync,
	}
	GitlabUserMessages = harness.Contract[
		[]harness.MergeRequest,
		[]harness.UserMessage,
	]{
		Domain:   "gitlab",
		Name:     "build-mr-user-messages",
		Version:  harness.V1,
		Tier:     harness.Admin,
		Priority: harness.Low,
		Strategy: harness.Sync,
	}
	SlackUserMessages = harness.Contract[
		[]harness.UserMessage,
		struct{},
	]{
		Domain:   "slack",
		Name:     "send-user-messages",
		Version:  harness.V1,
		Tier:     harness.Admin,
		Priority: harness.Low,
		Strategy: harness.Sync,
	}
	WorkflowGitlabNotifySlack = harness.Contract[string, struct{}]{
		Domain:   "coordinator",
		Name:     "fetch-mrs-and-send-user-messages",
		Version:  harness.V1,
		Tier:     harness.Admin,
		Priority: harness.Low,
		Strategy: harness.Sync,
	}
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
