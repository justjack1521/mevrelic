package mevrelic

import (
	"context"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type Context interface {
	ActualContext() context.Context
}

type Command interface {
	CommandName() string
}

type CommandHandler[C Command, R any] interface {
	Handle(ctx Context, cmd C) (R, error)
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

func (c commandHandlerWithNewRelic[C, R]) Handle(ctx Context, cmd C) (result R, err error) {

	var txn = newrelic.FromContext(ctx.ActualContext())

	var segment = txn.StartSegment(cmd.CommandName())

	defer func() {
		if err != nil {
			txn.NoticeError(err)
		}
		segment.End()
	}()

	return c.base.Handle(ctx, cmd)
}
