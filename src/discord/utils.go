package discord

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// SetupCloseHandler creates a handler that will catch SIGINT and SIGTERM signals
// and gracefully close the application
func SetupCloseHandler(cleanupFunc func() error) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nShutting down...")
		err := cleanupFunc()
		if err != nil {
			fmt.Printf("Error during cleanup: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
}

// CleanMentions removes Discord mentions from a message
func CleanMentions(content string, mentions []*User) string {
	for _, user := range mentions {
		content = strings.ReplaceAll(content, "<@"+user.ID+">", "")
		content = strings.ReplaceAll(content, "<@!"+user.ID+">", "")
	}
	return strings.TrimSpace(content)
}

// User represents a Discord user for mention cleaning
type User struct {
	ID string
}
