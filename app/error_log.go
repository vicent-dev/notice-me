package app

import (
	"log"
)

type serverErrorLogWriter struct{}

func (*serverErrorLogWriter) Write(p []byte) (int, error) {
	// ignore logs - it is handled by alog package
	return len(p), nil
}

func newServerErrorLog() *log.Logger {
	return log.New(&serverErrorLogWriter{}, "", 0)
}
