package harness

import (
	"context"
	"time"
)

type HandlerFunc[Input, Output any] func(context.Context, Input) (Output, error)

func (h HandlerFunc[Input, Output]) Handle(parent context.Context, input Input) (Output, error) {
	return h(parent, input)
}

type Strategy string

type Priority string

type Tier string

type Version string

const (
	Async Strategy = "async"
	Sync  Strategy = "sync"

	High Priority = "high"
	Low  Priority = "low"

	Admin   Tier = "admin"
	Public  Tier = "public"
	Session Tier = "session"

	V1 Version = "v1"
)

type Contract[Input, Output any] struct {
	Name    string
	Domain  string
	Version Version

	Priority Priority
	Strategy Strategy
	Tier     Tier
}

func (c Contract[Input, Output]) Queue() string {
	return string(c.Version) +
		"-" + c.Domain +
		"-" + string(c.Strategy) +
		"-" + string(c.Priority)
}

func (c Contract[Input, Output]) Workflow() string {
	return string(c.Version) +
		"-" + c.Domain +
		"-" + string(c.Name) +
		"-workflow"
}

func (c Contract[Input, Output]) Activity() string {
	return string(c.Version) +
		"-" + c.Domain +
		"-" + string(c.Name) +
		"-activity"
}

type Stream uint8

const (
	Stdin Stream = iota
	Stdout
	Stderr
)

type CommandEvent struct {
	Stream Stream
	Data   []byte
}

type InteractiveInput struct {
	Stream Stream
	Data   []byte
}

type SlackOAuthResponse struct {
	AccessToken                     string
	AppID                           string
	AuthedUserAccessToken           string
	AuthedUserExpiresIn             int32
	AuthedUserID                    string
	AuthedUserRefreshToken          string
	AuthedUserScope                 string
	AuthedUserTokenType             string
	BotUserID                       string
	EnterpriseID                    string
	EnterpriseName                  string
	Error                           string
	ExpiresIn                       int32
	IncomingWebhookChannel          string
	IncomingWebhookChannelID        string
	IncomingWebhookConfigurationURL string
	IncomingWebhookURL              string
	IsEnterpriseInstall             bool
	MetadataCursor                  string
	MetadataMessages                []string
	MetadataWarnings                []string
	OK                              bool
	RefreshToken                    string
	Scope                           string
	TeamID                          string
	TeamName                        string
	TokenType                       string
}

type MergeRequest struct {
	ID                          int
	IID                         int
	Jira                        string
	TargetBranch                string
	SourceBranch                string
	ProjectID                   int
	Title                       string
	State                       string
	Imported                    bool
	ImportedFrom                string
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	Upvotes                     int
	Downvotes                   int
	Author                      int32
	Assignee                    int
	Assignees                   []int32
	Reviewers                   []int32
	SourceProjectID             int
	TargetProjectID             int
	Labels                      []string
	Description                 string
	Draft                       bool
	MergeWhenPipelineSucceeds   bool
	DetailedMergeStatus         string
	MergeUser                   int32
	MergedAt                    time.Time
	MergeAfter                  time.Time
	PreparedAt                  time.Time
	ClosedBy                    int32
	ClosedAt                    time.Time
	SHA                         string
	MergeCommitSHA              string
	SquashCommitSHA             string
	UserNotesCount              int
	ShouldRemoveSourceBranch    bool
	ForceRemoveSourceBranch     bool
	AllowCollaboration          bool
	AllowMaintainerToPush       bool
	WebURL                      string
	ShortReference              string
	RelativeReference           string
	FullReference               string
	DiscussionLocked            bool
	Squash                      bool
	SquashOnMerge               bool
	TaskCount                   int
	TaskCompletedCount          int
	HasConflicts                bool
	BlockingDiscussionsResolved bool
}

type UserMessage struct {
	Email   string
	Message string
}

type GitlabUser struct {
	Email string
	ID    int32
}

type NamespacedObject struct {
	Namespace string
	ID        string
}

type Camera struct {
	UUID      string
	PanelHWID string
}

type CreateEmailRequest struct {
	UserID int64
	Email  string
}

type SavedEmail struct {
	ID     int64
	UserID int64
	Email  string
}
