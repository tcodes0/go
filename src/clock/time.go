package clock

import (
	"context"
	"time"
)

type Time struct {
	Location time.Location
}

var _ Nower = (*Time)(nil)

func (t *Time) Now() time.Time {
	return time.Now().In(&t.Location)
}

func (t *Time) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey, t)
}
