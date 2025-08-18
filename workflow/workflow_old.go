package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/logger"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type Start struct {
	QueueName string
	Client    client.Client
	Workflow  string
	Initial   interface{}
}

func (s *Start) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	we, err := s.Client.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		TaskQueue: config.Workflow.QueueName,
	}, s.Workflow, s.Initial)
	if err != nil {
		logger.Errorf(r.Context(), "could not execute workflow: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, "Started workflow %s", we.GetID())
}

type Wait[T any] struct {
	QueueName string
	Client    client.Client
	Workflow  string
	Initial   interface{}
}

func (t *Wait[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	we, err := t.Client.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		TaskQueue: config.Workflow.QueueName,
	}, t.Workflow, t.Initial)
	if err != nil {
		logger.Errorf(r.Context(), "could not execute workflow: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	var result T
	we.Get(r.Context(), &result)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err = enc.Encode(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RegisterActivity[A, B any](w worker.Worker, name string, handler harness.Handler[A, B]) {
	w.RegisterActivityWithOptions(
		handler.Handle,
		activity.RegisterOptions{
			Name: "activity-" + name,
		},
	)
}

func RegisterWorkflows(w worker.Worker) {
	w.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, input string) (struct{}, error) {
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: time.Second * 5,
			})
			var mrs []harness.MergeRequest
			var msgs []harness.UserMessage
			err := workflow.ExecuteActivity(ctx, "activity-gitlab-list-mrs", input).Get(ctx, &mrs)
			if err != nil {
				return struct{}{}, err
			}
			err = workflow.ExecuteActivity(ctx, "activity-gitlab-msgs", mrs).Get(ctx, &msgs)
			if err != nil {
				return struct{}{}, err
			}
			err = workflow.ExecuteActivity(ctx, "activity-slack-msgs", msgs).Get(ctx, nil)
			if err != nil {
				return struct{}{}, err
			}
			return struct{}{}, nil
		},
		workflow.RegisterOptions{
			Name: "slack-mrs",
		},
	)
	w.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, namespaces []string) ([]string, error) {
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: time.Second * 5,
			})
			var diff []string
			err := workflow.ExecuteActivity(ctx, "activity-diff-k8s-envs", namespaces).Get(ctx, &diff)
			if err != nil {
				return nil, err
			}
			return diff, nil
		},
		workflow.RegisterOptions{
			Name: "k8s-envs",
		},
	)
	w.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, camera harness.Camera) (string, error) {
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: time.Second * 5,
			})
			var result string
			err := workflow.ExecuteActivity(ctx, "activity-jwt", camera).Get(ctx, &result)
			if err != nil {
				return err.Error(), err
			}
			return result, nil
		},
		workflow.RegisterOptions{
			Name: "jwt",
		},
	)
	w.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, namespaces []string) ([]harness.NamespacedObject, error) {
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: time.Second * 5,
			})
			var diff []harness.NamespacedObject
			err := workflow.ExecuteActivity(ctx, "activity-list-k8s-diffs", namespaces).Get(ctx, &diff)
			if err != nil {
				return nil, err
			}
			return diff, nil
		},
		workflow.RegisterOptions{
			Name: "k8s-list",
		},
	)
	w.RegisterWorkflowWithOptions(
		func(ctx workflow.Context, input string) ([]harness.MergeRequest, error) {
			ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				StartToCloseTimeout: time.Second * 5,
			})
			var mrs []harness.MergeRequest
			err := workflow.ExecuteActivity(ctx, "port-forward", input).Get(ctx, &mrs)
			if err != nil {
				return nil, err
			}
			return mrs, nil
		},
		workflow.RegisterOptions{
			Name: "port-forward",
		},
	)
}
