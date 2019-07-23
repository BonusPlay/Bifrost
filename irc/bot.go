package irc

import (
	"crypto/tls"
	"github.com/BonusPlay/Bifrost/discord"
	. "github.com/BonusPlay/Bifrost/util"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	irc "github.com/thoj/go-ircevent"
)

func SetupBot() (irccon *irc.Connection) {
	cert, err := tls.LoadX509KeyPair(viper.GetString("irc.cert"), viper.GetString("irc.key"))
	CheckError("Failed to load certificates", err)

	irccon = irc.IRC("BonusPlay", "Bonus")
	irccon.RealName = "Adam Kliś"
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	err = irccon.Connect("chat.freenode.net:6697")
	CheckError("Failed to connect to Freenode", err)

	for _, channel := range viper.GetStringSlice("channels") {
		irccon.Join(channel)
	}

	return
}

func GetChannelId(dsession *discordgo.Session, event *irc.Event) (channelid string) {

	// status message
	if len(event.Nick) == 0 {
		channelid = discord.GetChannelByName(dsession, "status")

		if len(channelid) == 0 {
			log.Fatal("Failed to find status channel")
		}

		return
	}

	// channel message
	if event.Arguments[0][0] == '#' {
		channelName := event.Arguments[0][1:len(event.Arguments[0])]
		channelid = discord.GetChannelByName(dsession, channelName)

		if len(channelid) == 0 {
			channelid = discord.CreateDMChannel(dsession, channelName, "IRC-Channels").ID
		}

		return
	}

	// DM
	channelid = discord.GetChannelByName(dsession, event.Nick)
	if len(channelid) == 0 {
		channelid = discord.CreateDMChannel(dsession, event.Nick, "IRC-DMs").ID
	}

	return
}
