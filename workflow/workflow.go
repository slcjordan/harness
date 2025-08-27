package workflow

import (
	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/logger"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type WorkflowFunc[Input, Output any] func(workflow.Context, Input) (Output, error)

func Chain[A, B, C any](f WorkflowFunc[A, B], c harness.Contract[B, C]) WorkflowFunc[A, C] {
	next := Workflow(c)
	return func(ctx workflow.Context, input A) (C, error) {
		var result C
		b, err := f(ctx, input)
		if err != nil {
			return result, err
		}
		return next(ctx, b)
	}
}

func RegisterActivity[Input, Output any](c harness.Contract[Input, Output], w worker.Worker, handler harness.Handler[Input, Output]) {
	w.RegisterActivityWithOptions(
		handler.Handle,
		activity.RegisterOptions{
			Name: "activity-" + c.Name,
		},
	)
}

func Workflow[Input, Output any](c harness.Contract[Input, Output]) WorkflowFunc[Input, Output] {
	if c.Strategy == harness.Sync {
		return func(ctx workflow.Context, input Input) (Output, error) {
			var result Output
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: config.Workflow.ActivityTimeout,
				TaskQueue:           c.Queue(),
			})
			err := workflow.ExecuteActivity(ctx, "activity-"+c.Name, input).Get(ctx, &result)
			if err != nil {
				logger.Errorf(ctx, "error while running workflow %q: %s", c.Name, err)
			}
			return result, err
		}
	} else if c.Strategy == harness.Async {
		return func(ctx workflow.Context, input Input) (Output, error) {
			var result Output
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: config.Workflow.ActivityTimeout,
				TaskQueue:           c.Queue(),
			})
			workflow.ExecuteActivity(ctx, "activity-"+c.Name, input)
			return result, nil
		}
	}
	panic("unhandled strategy: " + c.Strategy)
}

func RegisterSynchronousWorkflow[Input, Output any](c harness.Contract[Input, Output], w worker.Worker) {
	w.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, input Input) (Output, error) {
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: config.Workflow.ActivityTimeout,
				TaskQueue:           c.Queue(),
			})
			var result Output
			err := workflow.ExecuteActivity(ctx, "activity-"+c.Name, input).Get(ctx, &result)
			if err != nil {
				logger.Errorf(ctx, "error while running workflow %q: %s", c.Name, err)
			}
			return result, err
		},
		workflow.RegisterOptions{
			Name: "workflow-" + c.Name,
		},
	)
}
