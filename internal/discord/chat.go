package discord

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// handleChatCommand handles the /chat command
func (b *DiscordBot) handleChatCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
	response, err := b.OpenAI.GenerateResponse(context.Background(), promptOption)
	if err != nil {
		fmt.Printf("Error generating response: %v\n", err)
		b.sendError(s, i, "AI Error", "Sorry, I couldn't process your request.")
		return
	}

	// Create an embed for the AI response
	embed := &discordgo.MessageEmbed{
		Title:       "AI Response",
		Description: response,
		Color:       0x00ff00,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Powered by OpenAI",
		},
	}

	// Check if response is too long for Discord embed description (max 4096 chars)
	if len(response) > 4000 {
		embed.Description = response[:4000] + "..."
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}); err != nil {
		fmt.Printf("Error editing interaction response: %v\n", err)
	}
}
