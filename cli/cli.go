package cli

import (
	"context"
	"net"
	"os"

	cli "github.com/urfave/cli/v2"

	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/config"
)

const EnvPrefix = "HARNESS_"

type Runner interface {
	Run(context.Context, []string) error
}

type RunnerFunc func(context.Context, []string) error

func (r RunnerFunc) Run(ctx context.Context, args []string) error {
	return r(ctx, args)
}

type Command struct {
	cmd *cli.Command
}

type Option func(*Command)

func NewCommand(name string, usage string, runner Runner, options ...Option) *Command {
	result := &Command{
		cmd: &cli.Command{
			Name:  name,
			Usage: usage,
			Action: func(c *cli.Context) error {
				return runner.Run(c.Context, c.Args().Slice())
			},
		},
	}
	for _, o := range options {
		o(result)
	}
	return result
}

func (c *Command) Subcommand(name string, usage string, runner Runner, options ...Option) {
	c.cmd.Subcommands = append(c.cmd.Subcommands, NewCommand(name, usage, runner, options...).cmd)
}

func (c *Command) Run(ctx context.Context, args []string) error {
	app := &cli.App{
		Name:   c.cmd.Name,
		Usage:  c.cmd.Usage,
		Flags:  c.cmd.Flags,
		Action: c.cmd.Action,
	}
	return app.RunContext(ctx, args)
}

func WithHTTPServerFlags(c *Command) {
	WithHTTPServerAddrFlag(c)
	WithHTTPServerFileRootFlag(c)
	WithHTTPServerKeyRootFlag(c)
}

func WithHTTPServerAddrFlag(c *Command) {
	c.cmd.Flags = append(
		c.cmd.Flags,
		&cli.StringFlag{
			Name:        "httpserver-addr",
			Usage:       "address to listen on",
			EnvVars:     []string{EnvPrefix + "HTTPSERVER_ADDR"},
			Destination: &config.HTTPServer.Addr,
			Value:       config.HTTPServer.Addr,
			Action: func(c *cli.Context, val string) error {
				_, err := net.ResolveTCPAddr("tcp", config.HTTPServer.Addr)
				return err
			},
		},
	)
}

func WithHTTPServerFileRootFlag(c *Command) {
	c.cmd.Flags = append(
		c.cmd.Flags,
		&cli.StringFlag{
			Name:        "httpserver-fileroot",
			EnvVars:     []string{EnvPrefix + "HTTPSERVER_FILEROOT"},
			Destination: &config.HTTPServer.FileRoot,
			Value:       config.HTTPServer.FileRoot,
			Action: func(c *cli.Context, val string) error {
				info, err := os.Stat(config.HTTPServer.FileRoot)
				if err != nil {
					return err
				}
				if !info.IsDir() {
					return harness.Errorf(harness.ErrInvalidSetting, "%v is not a directory", config.HTTPServer.FileRoot)
				}
				return nil
			},
		},
	)
}

func WithHTTPServerKeyRootFlag(c *Command) {
	c.cmd.Flags = append(
		c.cmd.Flags,
		&cli.StringFlag{
			Name:        "httpserver-keyroot",
			EnvVars:     []string{EnvPrefix + "HTTPSERVER_KEYROOT"},
			Destination: &config.HTTPServer.KeyRoot,
			Value:       config.HTTPServer.KeyRoot,
			Action: func(c *cli.Context, val string) error {
				info, err := os.Stat(config.HTTPServer.KeyRoot)
				if err != nil {
					return err
				}
				if !info.IsDir() {
					return harness.Errorf(harness.ErrInvalidSetting, "%v is not a directory", config.HTTPServer.KeyRoot)
				}
				return nil
			},
		},
	)
}

func WithSlackFlags(c *Command) {
	WithSlackClientIDFlag(c)
	WithSlackClientSecretFlag(c)
	WithSlackRedirectURLFlag(c)
}

func WithSlackClientIDFlag(c *Command) {
	c.cmd.Flags = append(
		c.cmd.Flags,
		&cli.StringFlag{
			Name:        "slack-client-id",
			Usage:       "address to listen on",
			EnvVars:     []string{EnvPrefix + "SLACK_CLIENT_ID"},
			Destination: &config.Slack.ClientID,
			Value:       config.Slack.ClientID,
		},
	)
}

func WithSlackClientSecretFlag(c *Command) {
	c.cmd.Flags = append(
		c.cmd.Flags,
		&cli.StringFlag{
			Name:        "slack-client-secret",
			EnvVars:     []string{EnvPrefix + "SLACK_CLIENT_SECRET"},
			Destination: &config.Slack.ClientSecret,
			Value:       config.Slack.ClientSecret,
		},
	)
}

func WithSlackRedirectURLFlag(c *Command) {
	c.cmd.Flags = append(
		c.cmd.Flags,
		&cli.StringFlag{
			Name:        "slack-redirect-url",
			EnvVars:     []string{EnvPrefix + "SLACK_REDIRECT_URL"},
			Destination: &config.Slack.RedirectURL,
			Value:       config.Slack.RedirectURL,
		},
	)
}

func WithPostgresDSNFlag(c *Command) {
	c.cmd.Flags = append(
		c.cmd.Flags,
		&cli.StringFlag{
			Name:        "postgres-dsn",
			EnvVars:     []string{EnvPrefix + "POSTGRES_DSN"},
			Destination: &config.Postgres.DSN,
			Value:       config.Postgres.DSN,
		},
	)
}
