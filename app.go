package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/BonusPlay/Bifrost/discord"
	ircbot "github.com/BonusPlay/Bifrost/irc"
	. "github.com/BonusPlay/Bifrost/util"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	irc "github.com/thoj/go-ircevent"
)

var dsession *discordgo.Session
var irccon *irc.Connection

func main() {

	err := setupConfig()
	CheckError("Failed to load config\n", err)

	irccon = ircbot.SetupBot()
	dsession = discord.SetupBot()

	irccon.AddCallback("PRIVMSG", onIrcMsg)
	dsession.AddHandler(onDiscordMsg)

	go irccon.Loop()
	err = dsession.Open()
	CheckError("Failed to start discord bot", err)

	log.Info("Bridge is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	irccon.Quit()
	_ = dsession.Close()
}

func setupConfig() (err error) {
	viper.SetConfigFile("bifrost.cfg")
	viper.SetConfigType("json")

	err = viper.ReadInConfig()
	CheckError("Failed to read config", err)

	return
}

// TODO: use discord webhooks to spoof message author
func onIrcMsg(event *irc.Event) {
	msg := &discordgo.MessageSend{
		Content: event.Message(),
		Embed: &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name: event.Nick,
			},
		},
	}

	channelid := ircbot.GetChannelId(dsession, event)
	_, err := dsession.ChannelMessageSendComplex(channelid, msg)
	CheckError("Failed to send discord message", err)
}

func onDiscordMsg(dssession *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == dssession.State.User.ID {
		return
	}

	channel, err := dsession.State.Channel(m.ChannelID)
	CheckError("Failed to fetch discord channel from message", err)
	parent, err := dsession.State.Channel(channel.ParentID)
	CheckError("Failed to fetch discord channel from message", err)

	var channelName string

	switch parent.Name {
	case "IRC-DMs":
		channelName = channel.Name
		break

	case "IRC-Channels":
		channelName = string('#') + channel.Name

	default:
		return
	}

	irccon.Privmsg(channelName, discord.SanitizeMsg(m))
}
