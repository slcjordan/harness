package workflow

import (
	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/logger"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func RegisterActivity[Input, Output any](c harness.Contract[Input, Output], w func(string) worker.Worker, handler harness.Handler[Input, Output]) {
	w(c.Queue()).RegisterActivityWithOptions(
		handler.Handle,
		activity.RegisterOptions{
			Name: c.Activity(),
		},
	)
}

func Step[Input, Output any](c harness.Contract[Input, Output]) func(workflow.Context, Input) (Output, error) {
	switch c.Strategy {
	case harness.Sync:
		return func(ctx workflow.Context, input Input) (Output, error) {
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: config.Workflow.ActivityTimeout,
				TaskQueue:           c.Queue(),
			})
			var result Output
			err := workflow.ExecuteActivity(ctx, c.Activity(), input).Get(ctx, &result)
			if err != nil {
				logger.Errorf(ctx, "error while running workflow %q: %s", c.Name, err)
			}
			return result, err
		}
	case harness.Async:
		return func(ctx workflow.Context, input Input) (Output, error) {
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: config.Workflow.ActivityTimeout,
				TaskQueue:           c.Queue(),
			})
			var result Output
			workflow.ExecuteActivity(ctx, c.Activity(), input)
			return result, nil
		}
	}
	return nil
}
