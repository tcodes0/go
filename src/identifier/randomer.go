package identifier

import (
	"context"
	"errors"
)

type ctxKeyIDGen struct{}

var contextKeyIDGenerator *ctxKeyIDGen = &ctxKeyIDGen{}

type Randomer interface {
	Random() string
}

func ContextWithRandomer(ctx context.Context, idgen Randomer) context.Context {
	return context.WithValue(ctx, contextKeyIDGenerator, idgen)
}

func ContextGetRandomer(ctx context.Context) (Randomer, error) {
	opts, ok := ctx.Value(contextKeyIDGenerator).(Randomer)
	if !ok {
		return nil, errors.New("key id generator: no value")
	}

	return opts, nil
}
