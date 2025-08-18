//go:build vivintgroup && activitycmd

package main

import (
	"go.temporal.io/sdk/worker"
)

func init() {
	WorkflowInit = append(WorkflowInit, func(w worker.Worker) error {
		VivintJWT.RegisterSynchronousWorkflow(w)
		return nil
	})
}
