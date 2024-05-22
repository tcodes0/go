package test

import (
	"log"
	"testing"

	"github.com/tcodes0/go/src/logging"
)

func TestLogger_Log(t *testing.T) {
	type fields struct {
		l         *log.Logger
		exit      func(code int)
		metadata  string
		level     logging.Level
		msgLevel  logging.Level
		color     bool
		calldepth int
	}
	type args struct {
		msg []interface{}
	}
	tests := []struct {
		name   string
		args   args
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.Create()
			logger.Log(tt.args.msg...)
		})
	}
}
