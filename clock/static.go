package clock

import (
	"context"
	time "time"
)

type Static struct {
	Source time.Time
}

var _ Nower = (*Static)(nil)

func (s Static) Now() time.Time {
	return s.Source
}

func (s Static) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, s)
}
