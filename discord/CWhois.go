package discord

import (
	. "github.com/BonusPlay/Bifrost/util"
	"github.com/bwmarrin/discordgo"
	"strings"
)

type CWhois struct
{}

func (_ CWhois) Description() string {
	return "list all users in channel"
}

func (_ CWhois) Run(msg *discordgo.Message) {
	content := msg.ContentWithMentionsReplaced()
	content = strings.TrimSpace(content)
	parts := strings.Split(content, " ")

	if len(parts) < 3 {
		_, err := Dsession.ChannelMessageSend(msg.ChannelID, "ERROR: Not enough arguments passed!")
		CheckError("Failed to send discord message", err)
		return
	}

	IrcSession.Whois(parts[2])
}