package dummy

import "time"

type Pet struct {
	Name string
	Age  int
}

type Human struct {
	Birthday time.Time
	Name     string
	LastName string
	Pets     []Pet
	Married  bool
}

func Dummy() string {
	return "v0.1.4"
}
