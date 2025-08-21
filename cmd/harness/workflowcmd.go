//go:build workflowcmd

package main

import (
	"context"
	"fmt"

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
			fmt.Println("workflow is running.")
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
			GitlabListMRs.RegisterSynchronousWorkflow(w)
			GitlabUserMessages.RegisterSynchronousWorkflow(w)
			SlackUserMessages.RegisterSynchronousWorkflow(w)

			// register complex workflows
			w.RegisterWorkflowWithOptions(
				workflow.Chain(
					workflow.Chain(
						GitlabListMRs.Workflow(),
						GitlabUserMessages),
					SlackUserMessages,
				),
				tworkflow.RegisterOptions{
					Name: workflow.Name(WorkflowGitlabNotifySlack),
				},
			)

			return nil
		}), WorkOptions...)
}
