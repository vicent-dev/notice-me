package main

import (
	"fmt"
	"os"
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

	args := os.Args[1:]

	if len(args) == 0 || args[0] == "generate" {
		plaintext, err := app.GenerateApiKeyCLI(db)
		if err != nil {
			fmt.Printf("Error generating API key: %v\n", err)
			return
		}
		fmt.Println(plaintext)
		return
	}

	if args[0] == "revoke" {
		if len(args) < 2 {
			fmt.Println("Usage: cli revoke <api-key-id>")
			return
		}
		err := app.RevokeApiKeyCLI(db, args[1])
		if err != nil {
			fmt.Printf("Error revoking API key: %v\n", err)
			return
		}
		fmt.Println("API key revoked successfully")
		return
	}

	fmt.Printf("Unknown command: %s\n", args[0])
	fmt.Println("Usage: cli [generate|revoke <api-key-id>]")
}
