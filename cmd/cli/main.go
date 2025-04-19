package main

import (
	"github.com/en-vee/alog"
	"notice-me-server/app"
)

func main() {
	s := app.NewServer()

	apiKey, err := s.GenerateApiKeyHandler()

	if err != nil {
		alog.Error("something went wrong: %v", err)
		return
	}

	alog.Info("Api key generated successfully: " + apiKey.Value)
}
