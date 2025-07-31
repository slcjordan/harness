package workflow

import (
	"context"
	"fmt"
	"net/http"

	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/logger"
	"go.temporal.io/sdk/client"
)

type Start struct {
	QueueName string
	Client    client.Client
}

func (s *Start) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	we, err := s.Client.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		TaskQueue: config.Workflow.QueueName,
	}, "slack-mrs", `{"msg": "helloworld"}`)
	if err != nil {
		logger.Errorf(r.Context(), "could not execute workflow: %s", err)
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, "Started workflow %s", we.GetID())
}
