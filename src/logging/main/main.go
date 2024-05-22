package main

import "github.com/tcodes0/go/src/logging"

func main() {
	l := logging.Create(logging.LInfo, 0 /*default*/, true)
	l.SetExit(func(code int) {})
	l.SetLevel(logging.LDebug)

	l.Log("starting up...")
	l.Warn().Logf("this is a warning")
	l.Log("shutting down...")
	l.Fatal("this is a fatal error")

	l.AppendMetadata("member", "thom")

	l.Error().Log("this is an error")
	l.Debug().Log("this is a debug message")

	l.AppendMetadata("email", "foo@bar.com")

	l.Log("done")
	l.WipeMetadata()

	l.Log("done")
	l.Fatalf("this is a fatal error: %s", "with a message")
}
