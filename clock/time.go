package clock

import (
	"context"
	"time"
)

// Time implements nower with runtime time and timezone.
type Time struct {
	Location time.Location
}

var _ Nower = (*Time)(nil)

// returns the current time in timezone.
func (t *Time) Now() time.Time {
	return time.Now().In(&t.Location)
}

// returns a context with Time.
func (t *Time) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, t)
}
