package mevrelic

import (
	"context"
	"github.com/newrelic/go-agent/v3/newrelic"
	"google.golang.org/grpc"
)

func ServerInterceptor(relic *NewRelic) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		var txn *newrelic.Transaction

		if relic != nil && relic.Application != nil {
			txn = relic.Application.StartTransaction(info.FullMethod)
			ctx = newrelic.NewContext(ctx, txn)
			defer txn.End()
		}

		h, err := handler(ctx, req)

		if err != nil {
			txn.NoticeError(err)
		}

		return h, err

	}
}
