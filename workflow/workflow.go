package workflow

import (
	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/logger"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type WorkflowFunc[A, B any] func(workflow.Context, A) (B, error)

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

func RegisterActivity[A, B any](c harness.Contract[A, B], w worker.Worker, handler harness.Handler[A, B]) {
	w.RegisterActivityWithOptions(
		handler.Handle,
		activity.RegisterOptions{
			Name: "activity-" + c.Name,
		},
	)
}

func Workflow[A, B any](c harness.Contract[A, B]) WorkflowFunc[A, B] {
	return func(ctx workflow.Context, input A) (B, error) {
		var result B
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: config.Workflow.ActivityTimeout,
			TaskQueue:           c.Queue,
		})
		err := workflow.ExecuteActivity(ctx, "activity-"+c.Name, input).Get(ctx, &result)
		if err != nil {
			logger.Errorf(ctx, "error while running workflow %q: %s", c.Name, err)
		}
		return result, err
	}
}

func RegisterSynchronousWorkflow[A, B any](c harness.Contract[A, B], w worker.Worker) {
	w.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, input A) (B, error) {
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: config.Workflow.ActivityTimeout,
				TaskQueue:           config.Workflow.QueueName,
			})
			var result B
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
