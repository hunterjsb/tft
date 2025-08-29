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
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		fmt.Printf("Error acknowledging interaction: %v\n", err)
		return
	}

	prompt := i.ApplicationCommandData().Options[0].StringValue()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := h.openAI.GenerateResponse(ctx, prompt)
	if err != nil {
		h.sendError(s, i, "AI Error", "Failed to generate response")
		return
	}

	if len(response) > 4000 {
		h.sendLongResponse(s, i, response)
	} else {
		h.sendResponse(s, i, response)
	}
}

// sendResponse sends a single embed response
func (h *ChatHandler) sendResponse(s *discordgo.Session, i *discordgo.InteractionCreate, response string) {
	embed := []*discordgo.MessageEmbed{{
		Title:       "AI Response",
		Description: response,
		Color:       0x00ff00,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Powered by OpenAI",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embed})
}

// sendLongResponse splits long responses into multiple embeds
func (h *ChatHandler) sendLongResponse(s *discordgo.Session, i *discordgo.InteractionCreate, response string) {
	chunks := h.chunkString(response, 3900)
	var embeds []*discordgo.MessageEmbed

	for i, chunk := range chunks {
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("AI Response (Part %d/%d)", i+1, len(chunks)),
			Description: chunk,
			Color:       0x00ff00,
		}

		if i == len(chunks)-1 {
			embed.Footer = &discordgo.MessageEmbedFooter{
				Text: "Powered by OpenAI",
			}
			embed.Timestamp = time.Now().Format(time.RFC3339)
		}

		embeds = append(embeds, embed)
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embeds})
}

// sendError sends an error embed
func (h *ChatHandler) sendError(s *discordgo.Session, i *discordgo.InteractionCreate, title, desc string) {
	embed := []*discordgo.MessageEmbed{{
		Title:       title,
		Description: desc,
		Color:       0xff0000,
	}}
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &embed})
}

// chunkString splits a string into smaller parts at word boundaries
func (h *ChatHandler) chunkString(s string, chunkSize int) []string {
	if len(s) <= chunkSize {
		return []string{s}
	}

	var chunks []string
	words := strings.Fields(s)
	currentChunk := ""

	for _, word := range words {
		if len(currentChunk)+len(word)+1 > chunkSize {
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
				currentChunk = word
			} else {
				// Word itself is too long, split by characters
				chunks = append(chunks, word[:chunkSize])
				currentChunk = word[chunkSize:]
			}
		} else {
			if currentChunk == "" {
				currentChunk = word
			} else {
				currentChunk += " " + word
			}
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}
