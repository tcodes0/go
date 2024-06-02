package clock

import (
	"context"
	time "time"
)

// Static implements nower with the same time returned every time.
type Static struct {
	Source time.Time
}

var _ Nower = (*Static)(nil)

// returns the same time every time.
func (s Static) Now() time.Time {
	return s.Source
}

// returns a context with Static.
func (s Static) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, s)
}
