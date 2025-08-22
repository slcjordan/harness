//go:build gitlabgroup && workcmd

package main

import (
	"context"

	ggitlab "gitlab.com/gitlab-org/api/client-go"
	"go.temporal.io/sdk/worker"

	"github.com/slcjordan/harness/cli"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/db"
	"github.com/slcjordan/harness/gitlab"
	"github.com/slcjordan/harness/workflow"
)

func init() {
	WorkInit = append(WorkInit, func(ctx context.Context, w worker.Worker) error {
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
		workflow.RegisterActivity(GitlabListMRs, w, listMRs)
		workflow.RegisterActivity(GitlabUserMessages, w, gitlabUserMessages)
		return nil
	})

	WorkOptions = append(WorkOptions, cli.WithPostgresDSNFlag, cli.WithGitlabTokenFlag)
}
