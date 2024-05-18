package clock

import time "time"

type Static struct {
	Source time.Time
}

var _ Nower = (*Static)(nil)

func (s Static) Now() time.Time {
	return s.Source
}
