package stdlogger

import (
	"fmt"
	"os"
)

type StdLogger struct{}

func (s *StdLogger) Info(msg string) {
	_, _ = fmt.Fprintln(os.Stdout, msg)
}

func (s *StdLogger) Error(msg string) {
	_, _ = fmt.Fprintln(os.Stderr, msg)
}

func (s *StdLogger) Errorf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}
