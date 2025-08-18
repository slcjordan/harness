//go:build vivintgroup && activitycmd

package main

import (
	"os"

	"github.com/slcjordan/harness/auth"
	"go.temporal.io/sdk/worker"
)

func init() {
	ActivityInit = append(ActivityInit, func(w worker.Worker) error {
		secret, err := os.ReadFile("/tmp/portal_test.key")
		if err != nil {
			return err
		}
		jwt := auth.JWT{
			Secret: secret,
		}
		VivintJWT.RegisterActivity(w, jwt)
		return nil
	})
}
