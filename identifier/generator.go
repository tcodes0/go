package identifier

import (
	"context"
	"errors"
)

type ContextKey struct{}

var contextKey = ContextKey{}

type Generator interface {
	Generate() string
	WithContext(ctx context.Context) context.Context
}

// retrieves a generator from the context.
func FromContext(ctx context.Context) (Generator, error) {
	rand, ok := ctx.Value(contextKey).(Generator)
	if !ok {
		return nil, errors.New("no value found in context")
	}

	return rand, nil
}
