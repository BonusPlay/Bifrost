package discord

import (
	. "github.com/BonusPlay/Bifrost/util"
	"github.com/bwmarrin/discordgo"
)

type CHelp struct
{}

func (_ CHelp) Description() string {
	return "prints this message"
}

func (_ CHelp) Run(msg *discordgo.Message) {
	reply := "commands:\n"
	for keyword, cmd := range Commands {
		reply += "`" + keyword + "` - " + cmd.Description() + "\n"
	}

	_, err := Dsession.ChannelMessageSend(msg.ChannelID, reply)
	CheckError("Failed to send message", err)
}