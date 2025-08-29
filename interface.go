package harness

import (
	"context"
)

type Handler[Input, Output any] interface {
	Handle(context.Context, Input) (Output, error)
}

type Daemon interface {
	Attach(id string)
	Unattach(id string)
}

type Notifier[Evt any] interface {
	Notify(id string, evt Evt)
}
