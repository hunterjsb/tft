package main

import (
	"fmt"
	"os"

	"github.com/hunterjsb/tft/src/discord"
	"github.com/hunterjsb/tft/src/dotenv"
)

func main() {
	// Load environment variables using custom dotenv module
	err := dotenv.LoadDefault()
	if err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
		fmt.Println("Continuing with environment variables from system...")
	}

	runDiscordBot()
}

func runDiscordBot() {
	fmt.Println("Starting Discord bot mode...")

	// Load Discord bot configuration
	config, err := discord.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	err = config.Validate()
	if err != nil {
		fmt.Printf("Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Create and start bot
	bot, err := discord.NewDiscordBot(config)
	if err != nil {
		fmt.Printf("Error creating bot: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting Discord bot...")
	err = bot.Start()
	if err != nil {
		fmt.Printf("Error starting bot: %v\n", err)
		os.Exit(1)
	}

	// Set up graceful shutdown
	discord.SetupCloseHandler(func() error {
		fmt.Println("Shutting down bot...")
		return bot.Stop()
	})

	// Block main goroutine indefinitely
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	select {}
}
