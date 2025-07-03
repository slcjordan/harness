package harness

type Listener[Req any] interface {
	Receive(id string, req Req)
	Done(id string)
}

type MaybeSender[Resp any] interface {
	MaybeSend(id string, resp Resp)
}
