package main

import (
	"notice-me-server/app"
)

func main() {
	s := app.NewServer()

	if err := s.Run(); err != nil {
		panic(err)
	}
}
