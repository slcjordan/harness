package slack

import (
	"context"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/config"
	"github.com/slcjordan/harness/logger"
)

var New = slack.New

type OAuthHandler struct {
	Client *http.Client
	Save   harness.Handler[harness.SlackOAuthResponse, struct{}]
}

func (o *OAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	ctx := r.Context()

	_, err := o.Handle(ctx, code)
	if err != nil {
		logger.Errorf(ctx, "could not save response: %s", err)
		return
	}
}

func (o *OAuthHandler) Handle(ctx context.Context, code string) (struct{}, error) {
	resp, err := slack.GetOAuthV2ResponseContext(ctx, o.Client, config.Slack.ClientID, config.Slack.ClientSecret, code, config.Slack.RedirectURL)
	if err != nil {
		return struct{}{}, err
	}
	_, err = o.Save.Handle(ctx, harness.SlackOAuthResponse{
		AccessToken:                     resp.AccessToken,
		AppID:                           resp.AppID,
		AuthedUserAccessToken:           resp.AuthedUser.AccessToken,
		AuthedUserExpiresIn:             int32(resp.AuthedUser.ExpiresIn),
		AuthedUserID:                    resp.AuthedUser.ID,
		AuthedUserRefreshToken:          resp.AuthedUser.RefreshToken,
		AuthedUserScope:                 resp.AuthedUser.Scope,
		AuthedUserTokenType:             resp.AuthedUser.TokenType,
		BotUserID:                       resp.BotUserID,
		EnterpriseID:                    resp.Enterprise.ID,
		EnterpriseName:                  resp.Enterprise.Name,
		Error:                           resp.Error,
		ExpiresIn:                       int32(resp.ExpiresIn),
		IncomingWebhookChannel:          resp.IncomingWebhook.Channel,
		IncomingWebhookChannelID:        resp.IncomingWebhook.ChannelID,
		IncomingWebhookConfigurationURL: resp.IncomingWebhook.ConfigurationURL,
		IncomingWebhookURL:              resp.IncomingWebhook.URL,
		IsEnterpriseInstall:             resp.IsEnterpriseInstall,
		MetadataCursor:                  resp.ResponseMetadata.Cursor,
		MetadataMessages:                resp.ResponseMetadata.Messages,
		MetadataWarnings:                resp.ResponseMetadata.Warnings,
		OK:                              resp.Ok,
		RefreshToken:                    resp.RefreshToken,
		Scope:                           resp.Scope,
		TeamID:                          resp.Team.ID,
		TeamName:                        resp.Team.Name,
		TokenType:                       resp.TokenType,
	})
	return struct{}{}, err
}

type PingEcho struct {
	Client       *slack.Client
	GetUserToken harness.Handler[string, string]
}

func (p *PingEcho) Handle(ctx context.Context, email string) (struct{}, error) {
	user, err := p.Client.GetUserByEmailContext(ctx, email)
	if err != nil {
		return struct{}{}, err
	}
	token, err := p.GetUserToken.Handle(ctx, user.ID)
	if err != nil {
		return struct{}{}, err
	}
	userClient := slack.New(token)
	channel, _, _, err := userClient.OpenConversationContext(ctx, &slack.OpenConversationParameters{
		ReturnIM: true,
		Users:    []string{user.ID},
	})
	if err != nil {
		return struct{}{}, err
	}
	_, _, err = userClient.PostMessage(channel.ID, slack.MsgOptionText("Hello from "+email, false))
	return struct{}{}, err
}

type UserMessages struct {
	Client       *slack.Client
	SenderEmail  string
	GetUserToken harness.Handler[string, string]
}

func (u *UserMessages) Handle(ctx context.Context, msgs []harness.UserMessage) (struct{}, error) {
	user, err := u.Client.GetUserByEmailContext(ctx, u.SenderEmail)
	if err != nil {
		return struct{}{}, err
	}
	token, err := u.GetUserToken.Handle(ctx, user.ID)
	if err != nil {
		return struct{}{}, err
	}
	userClient := slack.New(token)

	for _, msg := range msgs {
		dest, err := u.Client.GetUserByEmailContext(ctx, msg.Email)
		if err != nil {
			return struct{}{}, err
		}
		channel, _, _, err := userClient.OpenConversationContext(ctx, &slack.OpenConversationParameters{
			ReturnIM: true,
			Users:    []string{dest.ID},
		})
		if err != nil {
			return struct{}{}, err
		}
		_, _, err = userClient.PostMessage(channel.ID, slack.MsgOptionText(msg.Message, false))
	}
	return struct{}{}, err
}
