//go:build slackgroup && activitycmd

package main

import (
	"context"

	"github.com/slcjordan/harness/db"
	"github.com/slcjordan/harness/slack"
	"github.com/slcjordan/harness/workflow"
	"go.temporal.io/sdk/worker"
)

func init() {
	ActivityInit = append(ActivityInit, func(ctx context.Context, w worker.Worker) error {
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
}
