package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/db/sqlc"
)

func Connect(ctx context.Context) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, config.Postgres.DSN)
	if err != nil {
		panic(err)
	}
	return pool
}

type SaveSlackOAuthResponse struct {
	Pool *pgxpool.Pool
}

func nonNil[T any](val []T) []T {
	if val == nil {
		return make([]T, 0)
	}
	return val
}

func (s *SaveSlackOAuthResponse) Handle(ctx context.Context, input harness.SlackOAuthResponse) (struct{}, error) {
	arg := sqlc.InsertSlackOAuthResponseParams{
		AccessToken:                     input.AccessToken,
		AppID:                           input.AppID,
		AuthedUserAccessToken:           input.AuthedUserAccessToken,
		AuthedUserExpiresIn:             input.AuthedUserExpiresIn,
		AuthedUserID:                    input.AuthedUserID,
		AuthedUserRefreshToken:          input.AuthedUserRefreshToken,
		AuthedUserScope:                 input.AuthedUserScope,
		AuthedUserTokenType:             input.AuthedUserTokenType,
		BotUserID:                       input.BotUserID,
		EnterpriseID:                    input.EnterpriseID,
		EnterpriseName:                  input.EnterpriseName,
		Error:                           input.Error,
		ExpiresIn:                       input.ExpiresIn,
		IncomingWebhookChannel:          input.IncomingWebhookChannel,
		IncomingWebhookChannelID:        input.IncomingWebhookChannelID,
		IncomingWebhookConfigurationUrl: input.IncomingWebhookConfigurationURL,
		IncomingWebhookUrl:              input.IncomingWebhookURL,
		IsEnterpriseInstall:             input.IsEnterpriseInstall,
		MetadataCursor:                  input.MetadataCursor,
		MetadataMessages:                nonNil(input.MetadataMessages),
		MetadataWarnings:                nonNil(input.MetadataWarnings),
		Ok:                              input.OK,
		RefreshToken:                    input.RefreshToken,
		Scope:                           input.Scope,
		TeamID:                          input.TeamID,
		TeamName:                        input.TeamName,
		TokenType:                       input.TokenType,
	}
	queries := sqlc.New(s.Pool)
	return struct{}{}, queries.InsertSlackOAuthResponse(ctx, arg)
}

type Sanity struct {
	Pool *pgxpool.Pool
}

func (s *Sanity) Handle(ctx context.Context, _ struct{}) (struct{}, error) {
	return struct{}{}, sqlc.New(s.Pool).Sanity(ctx)
}

type GetBotAccessToken struct {
	Pool *pgxpool.Pool
}

func (b *GetBotAccessToken) Handle(ctx context.Context, _ struct{}) (string, error) {
	queries := sqlc.New(b.Pool)
	return queries.GetBotAccessToken(ctx)
}

type GetUserTokenByClientID struct {
	Pool *pgxpool.Pool
}

func (b *GetUserTokenByClientID) Handle(ctx context.Context, clientID string) (string, error) {
	queries := sqlc.New(b.Pool)
	return queries.GetUserTokenByClientID(ctx, clientID)
}

type GetGitlabUser struct {
	Pool *pgxpool.Pool
}

func (g *GetGitlabUser) Handle(ctx context.Context, gitlabUserID int32) (harness.GitlabUser, error) {
	queries := sqlc.New(g.Pool)
	email, err := queries.GetGitlabUser(ctx, gitlabUserID)
	if err != nil {
		return harness.GitlabUser{}, err
	}
	return harness.GitlabUser{
		ID:    gitlabUserID,
		Email: email,
	}, nil
}

type UpsertGitlabUser struct {
	Pool *pgxpool.Pool
}

func (u *UpsertGitlabUser) Handle(ctx context.Context, gitlabUser harness.GitlabUser) (struct{}, error) {
	queries := sqlc.New(u.Pool)
	return struct{}{}, queries.UpsertGitlabUser(ctx, sqlc.UpsertGitlabUserParams{
		GitlabUserID: gitlabUser.ID,
		Email:        gitlabUser.Email,
	})
}
