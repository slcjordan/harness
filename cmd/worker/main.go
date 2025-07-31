package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/db"
	"github.com/slcjordan/harness/gitlab"
	"github.com/slcjordan/harness/logger"
	"github.com/slcjordan/harness/slack"
	gitlabLib "gitlab.com/gitlab-org/api/client-go"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	logger.Init()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config.Workflow.Server = "workflow-server:7233"
	config.Workflow.QueueName = "harness-worker"

	c := cli.NewCommand("run", "run background worker", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
		fmt.Println("DSNDSN: ", config.Postgres.DSN)
		pool := db.Connect(ctx)
		defer pool.Close()
		sanity := &db.Sanity{
			Pool: pool,
		}
		_, err := sanity.Handle(ctx, struct{}{})
		if err != nil {
			return err
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
		gitlabClient, err := gitlabLib.NewClient(config.Gitlab.Token)
		if err != nil {
			return err
		}
		botAccessToken, err := (&db.GetBotAccessToken{
			Pool: pool,
		}).Handle(ctx, struct{}{})
		listMRsActivity := (&gitlab.ListMRs{
			Client: gitlabClient,
		}).Handle
		gitlabUserMessages := (&gitlab.UserMessages{
			Client: gitlabClient,
			CacheFetch: &db.GetGitlabUser{
				Pool: pool,
			},
			CacheSave: &db.UpsertGitlabUser{
				Pool: pool,
			},
		}).Handle
		slackUserMsgsActivity := (&slack.UserMessages{
			Client: slack.New(botAccessToken),
			GetUserToken: &db.GetUserTokenByClientID{
				Pool: pool,
			},
			SenderEmail: "jordan.crabtree@vivint.com",
		}).Handle

		w.RegisterActivityWithOptions(
			listMRsActivity,
			activity.RegisterOptions{
				Name: "gitlab-list-mrs",
			},
		)
		w.RegisterActivityWithOptions(
			gitlabUserMessages,
			activity.RegisterOptions{
				Name: "gitlab-msgs",
			},
		)
		w.RegisterActivityWithOptions(
			slackUserMsgsActivity,
			activity.RegisterOptions{
				Name: "slack-msgs",
			},
		)
		w.RegisterWorkflowWithOptions(
			func(ctx workflow.Context, input string) (struct{}, error) {
				ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
					StartToCloseTimeout: time.Second * 5,
				})
				var mrs []harness.MergeRequest
				var msgs []harness.UserMessage
				err := workflow.ExecuteActivity(ctx, "gitlab-list-mrs", input).Get(ctx, &mrs)
				if err != nil {
					return struct{}{}, err
				}
				err = workflow.ExecuteActivity(ctx, "gitlab-msgs", mrs).Get(ctx, &msgs)
				if err != nil {
					return struct{}{}, err
				}
				err = workflow.ExecuteActivity(ctx, "slack-msgs", msgs).Get(ctx, nil)
				if err != nil {
					return struct{}{}, err
				}
				return struct{}{}, nil
			},
			workflow.RegisterOptions{
				Name: "slack-mrs",
			},
		)

		return w.Run(worker.InterruptCh())
	}), cli.WithSlackFlags, cli.WithPostgresDSNFlag, cli.WithWorkflowFlags, cli.WithGitlabTokenFlag)

	err := c.Run(ctx, os.Args)
	if err != nil {
		logger.Infof(ctx, "application error: %s", err)
	}
}
