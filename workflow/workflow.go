package workflow

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/logger"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type WorkflowFunc[A, B any] func(workflow.Context, A) (B, error)

func Chain[A, B, C any](f WorkflowFunc[A, B], c Contract[B, C]) WorkflowFunc[A, C] {
	next := c.Workflow()
	return func(ctx workflow.Context, input A) (C, error) {
		var result C
		b, err := f(ctx, input)
		if err != nil {
			return result, err
		}
		return next(ctx, b)
	}
}

type Contract[A, B any] string

func (c Contract[A, B]) Name() string {
	return string(c)
}

func (c Contract[A, B]) Queue() string {
	return strings.Split(c.Name(), "-")[0]
}

func (c Contract[A, B]) RegisterActivity(w worker.Worker, handler harness.Handler[A, B]) {
	w.RegisterActivityWithOptions(
		handler.Handle,
		activity.RegisterOptions{
			Name: "activity-" + c.Name(),
		},
	)
}

func (c Contract[A, B]) Workflow() WorkflowFunc[A, B] {
	return func(ctx workflow.Context, input A) (B, error) {
		var result B
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout: config.Workflow.ActivityTimeout,
			TaskQueue:           c.Queue(),
		})
		err := workflow.ExecuteActivity(ctx, "activity-"+c.Name(), input).Get(ctx, &result)
		if err != nil {
			logger.Errorf(ctx, "error while running workflow %q: %s", c.Name(), err)
		}
		return result, err
	}
}

func (c Contract[A, B]) RegisterSynchronousWorkflow(w worker.Worker) {
	w.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, input A) (B, error) {
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: config.Workflow.ActivityTimeout,
				TaskQueue:           c.Queue(),
			})
			var result B
			err := workflow.ExecuteActivity(ctx, "activity-"+c.Name(), input).Get(ctx, &result)
			if err != nil {
				logger.Errorf(ctx, "error while running workflow %q: %s", c.Name(), err)
			}
			return result, err
		},
		workflow.RegisterOptions{
			Name: "workflow-" + c.Name(),
		},
	)
}

func (c Contract[A, B]) JSONHandler(wClient client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dec := json.NewDecoder(r.Body)
		var input A
		err := dec.Decode(&input)
		if err != nil {
			logger.Errorf(ctx, "could not decode %T from request body for %q workflow: %s", input, c.Name(), err)
			http.Error(w, err.Error(), 500)
			return
		}
		exe, err := wClient.ExecuteWorkflow(
			ctx, client.StartWorkflowOptions{}, "workflow-"+c.Name(), input,
		)
		if err != nil {
			logger.Errorf(ctx, "could not execute workflow %q: %s", c.Name(), err)
			http.Error(w, err.Error(), 500)
			return
		}
		var result B
		err = exe.Get(ctx, &result)
		if err != nil {
			logger.Errorf(ctx, "could not get workflow %q results: %s", c.Name(), err)
			http.Error(w, err.Error(), 500)
			return
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		err = enc.Encode(result)
		if err != nil {
			logger.Errorf(ctx, "could not encode workflow %q results: %s", c.Name(), err)
			http.Error(w, err.Error(), 500)
			return
		}
	}
}
