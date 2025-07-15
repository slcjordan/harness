package harness

import "context"

type Handler[Req, Resp any] interface {
	Handle(context.Context, Req) (Resp, error)
}

type Daemon interface {
	Attach(id string)
	Unattach(id string)
}

type Notifier[Evt any] interface {
	Notify(id string, evt Evt)
}
