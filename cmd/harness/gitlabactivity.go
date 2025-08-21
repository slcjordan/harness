//go:build gitlabgroup && activitycmd

package main

import (
	"context"

	ggitlab "gitlab.com/gitlab-org/api/client-go"
	"go.temporal.io/sdk/worker"

	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/db"
	"github.com/slcjordan/harness/gitlab"
)

func init() {
	ActivityInit = append(ActivityInit, func(ctx context.Context, w worker.Worker) error {
		pool, err := db.Connect(ctx)
		if err != nil {
			return err
		}
		gitlabClient, err := ggitlab.NewClient(config.Gitlab.Token)
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
		GitlabListMRs.RegisterActivity(w, listMRs)
		GitlabUserMessages.RegisterActivity(w, gitlabUserMessages)
		return nil
	})
}
