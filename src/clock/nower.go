package clock

import (
	"context"
	"errors"
	time "time"
)

type ContextKey struct{}

var contextKey = ContextKey{}

type Nower interface {
	Now() time.Time
	WithContext(ctx context.Context) context.Context
}

func FromContext(ctx context.Context) (Nower, error) {
	nower, ok := ctx.Value(contextKey).(Nower)
	if !ok {
		return nil, errors.New("no value found in context")
	}

	return nower, nil
}
