//go:build uicmd

package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.temporal.io/sdk/client"

	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	jhttp "github.com/slcjordan/harness/http"
	"github.com/slcjordan/harness/logger"
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
				subroute.Post("/"+WorkflowGitlabNotifySlack.Name, jhttp.JSONHandler(WorkflowGitlabNotifySlack, temporalClient))
			})
			logger.Infof(ctx, "serving at %q", config.HTTPServer.Addr)
			return http.ListenAndServe(config.HTTPServer.Addr, r)
		}), cli.WithWorkflowFlags, cli.WithHTTPServerFlags)
}
