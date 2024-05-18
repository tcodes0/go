package clock

import (
	"context"
	"errors"
	time "time"
)

type ctxKeyNow struct{}

var contextKeyNow *ctxKeyNow = &ctxKeyNow{}

type Nower interface {
	Now() time.Time
}

func ContextWithNower(ctx context.Context, n Nower) context.Context {
	return context.WithValue(ctx, contextKeyNow, n)
}

func ContextGetNower(ctx context.Context) (Nower, error) {
	nower, ok := ctx.Value(contextKeyNow).(Nower)
	if !ok {
		return nil, errors.New("key now: no value")
	}

	return nower, nil
}
