package json

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/slcjordan/harness"
	"github.com/slcjordan/harness/logger"
)

type InteractiveCommandEncoder struct {
	Command  harness.Handler[harness.InteractiveInput, struct{}]
	Listener harness.Notifier[[]byte]
}

func (p *InteractiveCommandEncoder) Handle(ctx context.Context, req []byte) (struct{}, error) {
	var input harness.InteractiveInput
	err := json.Unmarshal(req, &input)
	if err != nil {
		return struct{}{}, fmt.Errorf("handler error: %w", err)
	}
	return p.Command.Handle(ctx, input)
}

func (p *InteractiveCommandEncoder) Notify(id string, resp harness.CommandEvent) {
	output, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf(context.TODO(), "could not marshal command event: %s", err)
		return
	}
	p.Listener.Notify(id, output)
}
