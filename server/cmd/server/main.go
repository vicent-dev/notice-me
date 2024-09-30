package main

import (
	"notice-me-server/app"

	"github.com/en-vee/alog"
)

func main() {
	s := app.NewServer()

	if err := s.Run(); err != nil {
		alog.Error("server run error: " + err.Error())
	}
}
