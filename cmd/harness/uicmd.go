//go:build uicmd

package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/logger"
	"go.temporal.io/sdk/client"
)

func init() {
	config.Workflow.Namespace = "default"

	cmd.Subcommand(
		"ui", "run http server", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
			temporalClient, err := client.Dial(client.Options{
				HostPort:  config.Workflow.Server,
				Namespace: config.Workflow.Namespace,
			})
			if err != nil {
				return err
			}
			r := chi.NewRouter()
			r.Use(middleware.Logger)
			fs := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.FileServer(http.Dir(config.HTTPServer.FileRoot)).ServeHTTP(w, r)
			})
			r.Handle("/app/*", http.StripPrefix("/app/", fs))
			r.Route("/workflow", func(subroute chi.Router) {
				subroute.Post("/"+WorkflowGitlabNotifySlack.Name, JSONHandler(WorkflowGitlabNotifySlack, temporalClient))
			})
			logger.Infof(ctx, "serving at %q", config.HTTPServer.Addr)
			return http.ListenAndServe(config.HTTPServer.Addr, r)
		}), cli.WithWorkflowFlags, cli.WithHTTPServerFlags)
}

func JSONHandler[A, B any](c harness.Contract[A, B], wClient client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		dec := json.NewDecoder(r.Body)
		var input A
		err := dec.Decode(&input)
		if err != nil {
			logger.Errorf(ctx, "could not decode %T from request body for %q workflow: %s", input, c.Name, err)
			http.Error(w, err.Error(), 500)
			return
		}
		exe, err := wClient.ExecuteWorkflow(
			ctx, client.StartWorkflowOptions{
				TaskQueue: c.Queue,
			}, "workflow-"+c.Name, input,
		)
		if err != nil {
			logger.Errorf(ctx, "could not execute workflow %q: %s", c.Name, err)
			http.Error(w, err.Error(), 500)
			return
		}
		var result B
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
