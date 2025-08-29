package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// ChatHandler handles AI chat-related Discord commands
type ChatHandler struct {
	openAI *OpenAIClient
}

// NewChatHandler creates a new chat handler
func NewChatHandler(openAIToken string, maxTokens int, temperature float64) *ChatHandler {
	return &ChatHandler{
		openAI: NewOpenAIClient(openAIToken, maxTokens, temperature),
	}
}

// HandleChatCommand handles the /chat command
func (h *ChatHandler) HandleChatCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Acknowledge the interaction immediately
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		fmt.Printf("Error acknowledging interaction: %v\n", err)
		return
	}

	// Get the prompt option
	options := i.ApplicationCommandData().Options
	promptOption := options[0].StringValue()

	// Generate response from OpenAI
	response, err := h.openAI.GenerateResponse(context.Background(), promptOption)
	if err != nil {
		fmt.Printf("Error generating response: %v\n", err)
		errorEmbed := h.createErrorEmbed("AI Error", "Sorry, I couldn't process your request. Please try again later.")
		if _, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &errorEmbed,
		}); editErr != nil {
			fmt.Printf("Error editing interaction response: %v\n", editErr)
		}
		return
	}

	// Check if response is too long for Discord embed (max 4096 chars for description)
	if len(response) > 4000 {
		// Split into multiple embeds
		embeds := h.createLongResponseEmbeds(response)
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &embeds,
		}); err != nil {
			fmt.Printf("Error editing interaction response: %v\n", err)
		}
	} else {
		// Send as single embed
		embeds := h.createResponseEmbed(response)
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &embeds,
		}); err != nil {
			fmt.Printf("Error editing interaction response: %v\n", err)
		}
	}
}

// createResponseEmbed creates embed for AI response
func (h *ChatHandler) createResponseEmbed(response string) []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{
		{
			Title:       "🤖 AI Response",
			Description: response,
			Color:       0x00ff00,
			Timestamp:   time.Now().Format(time.RFC3339),
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Powered by OpenAI",
			},
		},
	}
}

// createLongResponseEmbeds creates multiple embeds for long AI responses
func (h *ChatHandler) createLongResponseEmbeds(response string) []*discordgo.MessageEmbed {
	var embeds []*discordgo.MessageEmbed

	// Split response into chunks (max 4000 chars per embed description)
	chunks := h.chunkString(response, 4000)

	for i, chunk := range chunks {
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("🤖 AI Response (Part %d of %d)", i+1, len(chunks)),
			Description: chunk,
			Color:       0x00ff00,
		}

		// Add footer only to the last embed
		if i == len(chunks)-1 {
			embed.Footer = &discordgo.MessageEmbedFooter{
				Text: "Powered by OpenAI",
			}
			embed.Timestamp = time.Now().Format(time.RFC3339)
		}

		embeds = append(embeds, embed)
	}

	return embeds
}

// createErrorEmbed creates error message embed
func (h *ChatHandler) createErrorEmbed(title, description string) []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{
		{
			Title:       fmt.Sprintf("❌ %s", title),
			Description: description,
			Color:       0xff0000,
			Timestamp:   time.Now().Format(time.RFC3339),
		},
	}
}

// chunkString splits a string into smaller parts
func (h *ChatHandler) chunkString(s string, chunkSize int) []string {
	if len(s) <= chunkSize {
		return []string{s}
	}

	var chunks []string

	// Try to split at line breaks when possible
	lines := strings.Split(s, "\n")
	currentChunk := ""

	for _, line := range lines {
		// If adding this line would exceed the chunk size, start a new chunk
		if len(currentChunk)+len(line)+1 > chunkSize {
			// If current chunk is not empty, add it to chunks
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
				currentChunk = ""
			}

			// If the line itself is too long, split it by characters
			if len(line) > chunkSize {
				for len(line) > 0 {
					if len(line) <= chunkSize {
						currentChunk = line
						break
					}

					chunks = append(chunks, line[:chunkSize])
					line = line[chunkSize:]
				}
			} else {
				currentChunk = line
			}
		} else {
			// Add line to current chunk
			if currentChunk == "" {
				currentChunk = line
			} else {
				currentChunk += "\n" + line
			}
		}
	}

	// Add the last chunk if it's not empty
	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}
