package mevrelic

import (
	"context"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type Command interface {
	CommandName() string
}

type CommandHandler[C Command, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

type commandHandlerWithNewRelic[C Command, R any] struct {
	relic *newrelic.Application
	base  CommandHandler[C, R]
}

func NewCommandDecoratorWithNewRelic[C Command, R any](nrl *newrelic.Application, handler CommandHandler[C, R]) CommandHandler[C, R] {
	return commandHandlerWithNewRelic[C, R]{
		relic: nrl,
		base:  handler,
	}
}

func (c commandHandlerWithNewRelic[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {

	txn := c.relic.StartTransaction(cmd.CommandName())
	var nrc = newrelic.NewContext(ctx, txn)

	defer func() {
		if err != nil {
			txn.NoticeError(err)
		}
		txn.End()
	}()

	return c.base.Handle(nrc, cmd)
}
