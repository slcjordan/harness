package main

// plswerk
import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/auth"
	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/db"
	"github.com/slcjordan/harness/gitlab"
	"github.com/slcjordan/harness/kube"
	"github.com/slcjordan/harness/logger"
	"github.com/slcjordan/harness/slack"
	wf "github.com/slcjordan/harness/workflow"
	gitlabLib "gitlab.com/gitlab-org/api/client-go"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	logger.Init()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config.Workflow.Server = "workflow-server:7233"
	config.Workflow.QueueName = "harness-worker"

	c := cli.NewCommand("run", "run background worker", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
		pool := db.Connect(ctx)
		defer pool.Close()
		sanity := &db.Sanity{
			Pool: pool,
		}
		err := errors.New("this is just an initial non-nil error state necessary for retry logic")
		for i := 0; i < 5 && err != nil; i++ {
			_, err = sanity.Handle(ctx, struct{}{})
			if err != nil {
				logger.Infof(ctx, "waiting for db connection to become healthy: %s", err)
				time.Sleep(3 * time.Second)
			}
		}
		if err != nil {
			return err
		}

		gitlabClient, err := gitlabLib.NewClient(config.Gitlab.Token)
		if err != nil {
			return err
		}
		listMRs := &gitlab.ListMRs{
			Client: gitlabClient,
		}
		gitlabUserMessages := &gitlab.UserMessages{
			Client: gitlabClient,
			CacheFetch: &db.GetGitlabUser{
				Pool: pool,
			},
			CacheSave: &db.UpsertGitlabUser{
				Pool: pool,
			},
		}
		secret, err := os.ReadFile("/tmp/portal_test.key")
		if err != nil {
			panic(err)
		}
		jwt := auth.JWT{
			Secret: secret,
		}

		cfg, err := clientcmd.BuildConfigFromFlags("", "/home/jcrabtree/.kube/config")
		if err != nil {
			panic(err.Error())
		}
		clientset, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return err
		}
		diffK8sEnvs := &kube.Differ{
			Clientset: clientset,
		}
		listK8sEnvs := &kube.Lister{
			Clientset: clientset,
		}
		portForward := &kube.PortForwardDeploy[string, []harness.MergeRequest]{
			Config:    cfg,
			Clientset: clientset,
			Name:      "panel",
			Namespace: "jcrabtree",
			Ports:     []string{"4321:4321"},
			Handler:   listMRs,
		}

		botAccessToken, err := (&db.GetBotAccessToken{
			Pool: pool,
		}).Handle(ctx, struct{}{})
		slackUserMsgsActivity := &slack.UserMessages{
			Client: slack.New(botAccessToken),
			GetUserToken: &db.GetUserTokenByClientID{
				Pool: pool,
			},
			SenderEmail: "jordan.crabtree@vivint.com",
		}

		workerClient, err := client.Dial(client.Options{
			HostPort:  config.Workflow.Server,
			Namespace: "default",
		})
		if err != nil {
			return err
		}
		defer workerClient.Close()
		w := worker.New(workerClient, config.Workflow.QueueName, worker.Options{})
		wf.RegisterActivity(w, "jwt", jwt)
		wf.RegisterActivity(w, "gitlab-list-mrs", listMRs)
		wf.RegisterActivity(w, "gitlab-msgs", gitlabUserMessages)
		wf.RegisterActivity(w, "slack-msgs", slackUserMsgsActivity)
		wf.RegisterActivity(w, "diff-k8s-envs", diffK8sEnvs)
		wf.RegisterActivity(w, "list-k8s-diffs", listK8sEnvs)
		wf.RegisterActivity(w, "port-forward", portForward)
		wf.RegisterWorkflows(w)

		return w.Run(worker.InterruptCh())
	}), cli.WithSlackFlags, cli.WithPostgresDSNFlag, cli.WithWorkflowFlags, cli.WithGitlabTokenFlag)

	err := c.Run(ctx, os.Args)
	if err != nil {
		logger.Infof(ctx, "application error: %s", err)
	}
}
