package tests

import (
	"github.com/BonusPlay/Bifrost/discord"
	"github.com/bwmarrin/discordgo"
	"testing"
)

func TestSanitizeMsg(t *testing.T) {
	param := discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content: "Example <@123456789012345678> message with stuff <text> <:emote:123456789012345678",
		},
	}

	_ = discord.SanitizeMsg(&param)
}