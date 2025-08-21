//go:build workflowcmd

package main

import (
	"context"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	tworkflow "go.temporal.io/sdk/workflow"

	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/workflow"
)

func init() {
	config.Workflow.QueueName = "harness-worker"

	cmd.Subcommand(
		"workflow", "run temporal workflow worker", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
			workerClient, err := client.Dial(client.Options{
				HostPort:  config.Workflow.Server,
				Namespace: "default",
			})
			if err != nil {
				return err
			}
			defer workerClient.Close()
			w := worker.New(workerClient, config.Workflow.QueueName, worker.Options{})

			// register single-activity workflows
			workflow.RegisterSynchronousWorkflow(GitlabListMRs, w)
			workflow.RegisterSynchronousWorkflow(GitlabUserMessages, w)
			workflow.RegisterSynchronousWorkflow(SlackUserMessages, w)

			// register complex workflows
			w.RegisterWorkflowWithOptions(
				workflow.Chain(
					workflow.Chain(
						workflow.Workflow(GitlabListMRs),
						GitlabUserMessages),
					SlackUserMessages,
				),
				tworkflow.RegisterOptions{
					Name: WorkflowGitlabNotifySlack.Name,
				},
			)

			return nil
		}), WorkOptions...)
}
