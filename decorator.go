package mevrelic

import (
	"context"
	"github.com/justjack1521/mevrpc"
	"github.com/newrelic/go-agent/v3/newrelic"
	uuid "github.com/satori/go.uuid"
)

type Context interface {
	context.Context
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

	var txn = newrelic.FromContext(ctx)
	var segment = txn.StartSegment(cmd.CommandName())

	var user = mevrpc.UserIDFromContext(ctx)
	if user != uuid.Nil {
		txn.AddAttribute("user.id", user.String())
	}

	var player = mevrpc.PlayerIDFromContext(ctx)
	if player != uuid.Nil {
		txn.AddAttribute("player.id", user.String())
	}

	defer func() {
		if err != nil {
			txn.NoticeError(err)
		}
		segment.End()
	}()

	return c.base.Handle(ctx, cmd)
}
