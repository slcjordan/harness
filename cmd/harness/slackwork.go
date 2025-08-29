//go:build slackgroup && workcmd

package main

import (
	"context"

	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/db"
	"github.com/slcjordan/harness/scheduler"
	"github.com/slcjordan/harness/slack"
	"github.com/slcjordan/harness/workflow"
	"go.temporal.io/sdk/worker"
)

func init() {
	RegisterWorkflowHook(func(ctx context.Context, w func(string) worker.Worker, p *scheduler.Priority) error {
		pool, err := db.Connect(ctx)
		if err != nil {
			return err
		}
		botAccessToken, err := (&db.GetBotAccessToken{
			Pool: pool,
		}).Handle(ctx, struct{}{})
		if err != nil {
			return err
		}
		slackUserMsgsActivity := &slack.UserMessages{
			Client: slack.New(botAccessToken),
			GetUserToken: &db.GetUserTokenByClientID{
				Pool: pool,
			},
			SenderEmail: "jordan.crabtree@vivint.com",
		}
		workflow.RegisterActivity(SlackUserMessages, w, slackUserMsgsActivity)
		return nil
	})

	WorkOptions = append(WorkOptions, cli.WithSlackFlags)
}
