package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"

	"github.com/coder/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/db"
	"github.com/slcjordan/harness/exec"
	"github.com/slcjordan/harness/json"
	"github.com/slcjordan/harness/logger"
	"github.com/slcjordan/harness/pki"
	"github.com/slcjordan/harness/slack"
	"github.com/slcjordan/harness/workflow"
	"github.com/slcjordan/harness/ws"
	"go.temporal.io/sdk/client"
)

func mustLoadTLSConfig() (*http.Client, *tls.Config) {
	caKeyfile := filepath.Join(config.HTTPServer.KeyRoot, "ca_key.pem")
	caCertfile := filepath.Join(config.HTTPServer.KeyRoot, "ca_cert.pem")
	eeKeyfile := filepath.Join(config.HTTPServer.KeyRoot, "ee_key.pem")

	provider, err := pki.NewProvider(caKeyfile, caCertfile, eeKeyfile)
	if err != nil {
		panic(err)
	}
	certPool, err := provider.CACertPool()
	if err != nil {
		panic(err)
	}
	tlsConfig, err := provider.TLSConfig("localhost", []string{"My Server"})
	if err != nil {
		panic(err)
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: certPool},
		},
	}, tlsConfig
}

func main() {
	logger.Init()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config.Workflow.Server = "workflow-server:7233"
	config.Workflow.QueueName = "harness-worker"

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	enc := &json.InteractiveCommandEncoder{}
	// daemon := exec.StartInteractive(ctx, enc, "portal-tester", "-portalHost", "portal", "-cameraJWT", "", "-uuid", "a6911eb4-c4be-4986-adec-584e9ae47a66", "-run", "sendDetectionEvents", "-input-file", "/dev/stdin")
	daemon := exec.StartInteractive(ctx, enc, "cat")
	enc.Command = daemon
	ws := &ws.Server{
		Conns:    make(map[string]*websocket.Conn),
		Model:    daemon,
		Command:  enc,
		Listener: enc,
	}
	enc.Listener = ws
	tlsClient, tlsConfig := mustLoadTLSConfig()
	slackOAuth := &slack.OAuthHandler{
		Client: tlsClient,
	}

	c := cli.NewCommand("serve", "run http server", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
		pool := db.Connect(ctx)
		defer pool.Close()
		saveSlackOAuthResponse := &db.SaveSlackOAuthResponse{
			Pool: pool,
		}
		slackOAuth.Save = saveSlackOAuthResponse
		sanity := &db.Sanity{
			Pool: pool,
		}
		_, err := sanity.Handle(ctx, struct{}{})
		if err != nil {
			return err
		}

		temporalClient, err := client.Dial(client.Options{
			HostPort:  config.Workflow.Server,
			Namespace: "default",
		})
		if err != nil {
			return err
		}
		gitStart := &workflow.Start{
			Client:    temporalClient,
			QueueName: config.Workflow.QueueName,
			Initial:   "",
			Workflow:  "slack-mrs",
		}
		k8sWait := &workflow.Wait[[]string]{
			Client:    temporalClient,
			QueueName: config.Workflow.QueueName,
			Initial:   []string{"smoke", "jcrabtree"},
			Workflow:  "k8s-envs",
		}
		k8sListWait := &workflow.Wait[[]harness.NamespacedObject]{
			Client:    temporalClient,
			QueueName: config.Workflow.QueueName,
			Initial:   []string{"smoke", "jcrabtree"},
			Workflow:  "k8s-list",
		}
		jwtWait := &workflow.Wait[string]{
			Client:    temporalClient,
			QueueName: config.Workflow.QueueName,
			Initial: harness.Camera{
				UUID:      "676959d4-f291-44e4-b6c0-377299238597",
				PanelHWID: "HW4-17068E84BD4",
			},
			Workflow: "jwt",
		}
		portForwardWait := &workflow.Wait[[]harness.MergeRequest]{
			Client:    temporalClient,
			QueueName: config.Workflow.QueueName,
			Initial:   "",
			Workflow:  "port-forward",
		}

		fs := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.FileServer(http.Dir(config.HTTPServer.FileRoot)).ServeHTTP(w, r)
		})
		r.Handle("/app/*", http.StripPrefix("/app/", fs))
		r.Handle("/app/ws/*", http.StripPrefix("/app/ws/", ws))
		r.Handle("/slack/oauth/callback", slackOAuth)
		r.Handle("/git", gitStart)
		r.Handle("/k8s", k8sWait)
		r.Handle("/k8s-list", k8sListWait)
		r.Handle("/jwt", jwtWait)
		r.Handle("/port-forward", portForwardWait)
		server := http.Server{
			Addr:    config.HTTPServer.Addr,
			Handler: r,
		}
		server.TLSConfig = tlsConfig
		logger.Infof(ctx, "listening at %q", config.HTTPServer.Addr)
		return http.ListenAndServe(config.HTTPServer.Addr, r)
	}), cli.WithHTTPServerFlags, cli.WithSlackFlags, cli.WithPostgresDSNFlag, cli.WithWorkflowFlags)

	err := c.Run(ctx, os.Args)
	if err != nil {
		logger.Infof(ctx, "application error: %s", err)
	}
}
