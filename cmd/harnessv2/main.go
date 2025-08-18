package main

import (
	"context"
	"os"

	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/logger"
	"github.com/slcjordan/harness/workflow"
)

func init() {
	config.Workflow.Server = "workflow-server:7233"
}

var WorkOptions []cli.Option
var ServeOptions []cli.Option

const (
	VivintJWT     workflow.Contract[harness.Camera, string]               = "vivint-jwt"
	GitlabListMRs workflow.Contract[string, harness.MergeRequest]         = "gitlab-list-mrs"
	UserMessages  workflow.Contract[[]harness.UserMessage, struct{}]      = "gitlab-user-messages"
	K8sDiffEnvs   workflow.Contract[[]string, []string]                   = "k8s-diff-envs"
	K8sListEnvs   workflow.Contract[[]string, []harness.NamespacedObject] = "k8s-list-envs"
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
