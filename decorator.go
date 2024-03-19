package mevrelic

import (
	"context"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type Command interface {
	Name() string
}

type CommandHandler[C Command] interface {
	Handle(ctx context.Context, cmd C) error
}

type QueryHandler[C Command, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

type commandHandlerWithNewRelic[C Command] struct {
	relic *newrelic.Application
	base  CommandHandler[C]
}

func NewCommandDecoratorWithNewRelic[C Command](nrl *newrelic.Application, handler CommandHandler[C]) CommandHandler[C] {
	return commandHandlerWithNewRelic[C]{
		relic: nrl,
		base:  handler,
	}
}

func (c commandHandlerWithNewRelic[C]) Handle(ctx context.Context, cmd C) (err error) {

	txn := c.relic.StartTransaction(cmd.Name())
	var nrc = newrelic.NewContext(ctx, txn)

	defer func() {
		if err != nil {
			txn.NoticeError(err)
		}
		txn.End()
	}()

	return c.base.Handle(nrc, cmd)
}

type queryHandlerWithNewRelic[C Command, R any] struct {
	relic *newrelic.Application
	base  QueryHandler[C, R]
}

func NewQueryDecoratorWithNewRelic[C Command, R any](nrl *newrelic.Application, handler QueryHandler[C, R]) QueryHandler[C, R] {
	return queryHandlerWithNewRelic[C, R]{
		relic: nrl,
		base:  handler,
	}
}

func (c queryHandlerWithNewRelic[C, R]) Handle(ctx context.Context, cmd C) (response R, err error) {

	txn := c.relic.StartTransaction(cmd.Name())
	var nrc = newrelic.NewContext(ctx, txn)

	defer func() {
		if err != nil {
			txn.NoticeError(err)
		}
		txn.End()
	}()

	return c.base.Handle(nrc, cmd)
}
