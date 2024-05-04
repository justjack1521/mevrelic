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

type CommandHandler[CTX Context, C Command, R any] interface {
	Handle(ctx CTX, cmd C) (R, error)
}

type commandHandlerWithNewRelic[CTX Context, C Command, R any] struct {
	relic *newrelic.Application
	base  CommandHandler[CTX, C, R]
}

func NewCommandDecoratorWithNewRelic[CTX Context, C Command, R any](nrl *newrelic.Application, handler CommandHandler[CTX, C, R]) CommandHandler[CTX, C, R] {
	return commandHandlerWithNewRelic[CTX, C, R]{
		relic: nrl,
		base:  handler,
	}
}

func (c commandHandlerWithNewRelic[CTX, C, R]) Handle(ctx CTX, cmd C) (result R, err error) {

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
