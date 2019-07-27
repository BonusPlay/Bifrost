package util

import (
	"github.com/bwmarrin/discordgo"
	"github.com/thoj/go-ircevent"
	"log"
)

// this is turbo hacky, but avoids import cycle that I'm too lazy to fix properly
var Dsession *discordgo.Session
var IrcSession *irc.Connection

func CheckError(msg string, err error) {
	if err != nil {
		log.Fatal(msg, '\n', err)
	}
}