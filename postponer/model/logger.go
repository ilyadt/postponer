package model

type Logger interface {
	Info(msg string)
	Error(msg string)
	Errorf(format string, a ...interface{})
}
