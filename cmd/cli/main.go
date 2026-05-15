package main

import (
	"fmt"
	"notice-me-server/app"
	"notice-me-server/pkg/config"
)

func main() {
	cfg := config.LoadConfig()

	db, err := app.InitDB(cfg)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return
	}

	apiKey, err := app.GenerateApiKeyCLI(db)
	if err != nil {
		fmt.Printf("Error generating API key: %v\n", err)
		return
	}

	fmt.Println(apiKey.Value)
}
