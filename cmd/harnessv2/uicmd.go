//go:build uicmd

package main

import (
	"context"

	"github.com/slcjordan/harness/cli"
)

func init() {
	cmd.Subcommand(
		"serve-ui", "run http server", cli.RunnerFunc(func(ctx context.Context, _ []string) error {
			return nil
		}), ServeOptions...)
}
