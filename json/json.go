package json

import (
	"encoding/json"

	"github.com/slcjordan/harness"
)

type InteractiveCommandEncoder struct {
	Listener    harness.Listener[harness.InteractiveExecInput]
	MaybeSender harness.MaybeSender[[]byte]
}

func (p *InteractiveCommandEncoder) Receive(id string, req []byte) {
	var input harness.InteractiveExecInput
	err := json.Unmarshal(req, &input)
	if err != nil {
		// TODO log
		return
	}
	p.Listener.Receive(id, input)
}

func (p *InteractiveCommandEncoder) Send(id string, resp harness.InteractiveExecOutput) {
	output, err := json.Marshal(resp)
	if err != nil {
		// TODO log
		return
	}
	p.MaybeSender.MaybeSend(id, output)
}
