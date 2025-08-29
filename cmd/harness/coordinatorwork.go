//go:build coordinatorgroup && workcmd

package main

import (
	"context"

	"go.temporal.io/sdk/worker"
	workflowlib "go.temporal.io/sdk/workflow"

	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/scheduler"
	"github.com/slcjordan/harness/workflow"
)

func registerWorkflow[Input, Output any](
	w func(string) worker.Worker,
	p *scheduler.Priority,
	c harness.Contract[Input, Output],
	handler scheduler.HandlerFunc[workflowlib.Context, Input, Output],
) {

	w(c.Queue()).RegisterWorkflowWithOptions(
		scheduler.Wrap(c, handler, p),
		workflowlib.RegisterOptions{Name: c.Name},
	)
}

func init() {
	config.Workflow.QueueName = "harness-worker"

	RegisterWorkflowHook(func(ctx context.Context, w func(string) worker.Worker, p *scheduler.Priority) error {
		// register single-activity workflows
		registerWorkflow(w, p, GitlabListMRs, workflow.Step(GitlabListMRs))
		registerWorkflow(w, p, GitlabUserMessages, workflow.Step(GitlabUserMessages))
		registerWorkflow(w, p, SlackUserMessages, workflow.Step(SlackUserMessages))

		// register complex workflows
		registerWorkflow(w, p, WorkflowGitlabNotifySlack,
			func(ctx workflowlib.Context, input string) (struct{}, error) {
				mrs, err := workflow.Step(GitlabListMRs)(ctx, input)
				if err != nil {
					return struct{}{}, err
				}
				gmsg, err := workflow.Step(GitlabUserMessages)(ctx, mrs)
				if err != nil {
					return struct{}{}, err
				}
				return workflow.Step(SlackUserMessages)(ctx, gmsg)
			})
		return nil
	})
}
