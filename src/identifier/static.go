package identifier

import (
	"context"
	"strconv"
)

type StaticGenerator struct {
	Prefix string
	Count  int
}

var _ Generator = (*StaticGenerator)(nil)

// generates a new static identifier.
func (static *StaticGenerator) Generate() string {
	id := static.Prefix + "-" + strconv.Itoa(static.Count)
	static.Count += 1

	return id
}

// returns a new context with the static generator.
func (static *StaticGenerator) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, static)
}
