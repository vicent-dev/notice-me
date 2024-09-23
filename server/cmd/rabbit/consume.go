package main

import "notice-me-server/app"

func main() {
	s := app.NewServer()
	s.DeclareQueues()
	s.RunConsummers()
}
