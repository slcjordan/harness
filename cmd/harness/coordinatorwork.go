//go:build coordinatorgroup && workcmd

package main

import (
	"context"

	"go.temporal.io/sdk/worker"
	tworkflow "go.temporal.io/sdk/workflow"

	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/workflow"
)

func init() {
	config.Workflow.QueueName = "harness-worker"

	WorkInit = append(WorkInit, func(ctx context.Context, w worker.Worker) error {
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

		return w.Run(worker.InterruptCh())
	})
}
