package http

import (
	"encoding/json"
	"net/http"

	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/logger"
	"go.temporal.io/sdk/client"
)

func JSONHandler[Input, Output any](c harness.Contract[Input, Output], wClient client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dec := json.NewDecoder(r.Body)
		var input Input
		err := dec.Decode(&input)
		if err != nil {
			logger.Errorf(ctx, "could not decode %T from request body for %q workflow: %s", input, c.Name, err)
			http.Error(w, err.Error(), 500)
			return
		}
		exe, err := wClient.ExecuteWorkflow(
			ctx, client.StartWorkflowOptions{
				TaskQueue: c.Queue(),
			}, "workflow-"+c.Name, input,
		)
		if err != nil {
			logger.Errorf(ctx, "could not execute workflow %q: %s", c.Name, err)
			http.Error(w, err.Error(), 500)
			return
		}
		var result Output
		err = exe.Get(ctx, &result)
		if err != nil {
			logger.Errorf(ctx, "could not get workflow %q results: %s", c.Name, err)
			http.Error(w, err.Error(), 500)
			return
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		err = enc.Encode(result)
		if err != nil {
			logger.Errorf(ctx, "could not encode workflow %q results: %s", c.Name, err)
			http.Error(w, err.Error(), 500)
			return
		}
	}
}
