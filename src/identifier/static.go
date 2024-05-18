package identifier

import "strconv"

type StaticGenerator struct {
	Source string

	i int
}

var _ Randomer = (*StaticGenerator)(nil)

func (s *StaticGenerator) Random() string {
	s.i += 1

	return s.Source + "-" + strconv.Itoa(s.i)
}
