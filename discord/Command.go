package discord

import "github.com/bwmarrin/discordgo"

type Command interface {
	Description() string
	Run(msg *discordgo.Message)
}

//func Split(msg *discordgo.Message) (cmd, args string) {
//
//}