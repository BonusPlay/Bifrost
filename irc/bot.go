package irc

import (
	"crypto/tls"
	"github.com/BonusPlay/Bifrost/discord"
	. "github.com/BonusPlay/Bifrost/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	irc "github.com/thoj/go-ircevent"
	"strings"
)

func SetupBot() {
	cert, err := tls.LoadX509KeyPair(viper.GetString("irc.cert"), viper.GetString("irc.key"))
	CheckError("Failed to load certificates", err)

	IrcSession = irc.IRC(viper.GetString("irc.nick"), viper.GetString("irc.user"))
	if viper.IsSet("irc.real_name") {
		IrcSession.RealName = viper.GetString("irc.real_name")
	}
	IrcSession.UseTLS = true
	IrcSession.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	err = IrcSession.Connect("chat.freenode.net:6697")
	CheckError("Failed to connect to Freenode", err)

	IrcSession.AddCallback("PRIVMSG", onIrcMsg)
	IrcSession.AddCallback("001", onIrcConnected)
}

func GetChannelId(event *irc.Event) (channelid string) {

	// status message
	if len(event.Nick) == 0 {
		channelid = discord.GetChannelByName("status")

		if len(channelid) == 0 {
			log.Fatal("Failed to find status channel")
		}

		return
	}

	// channel message
	if event.Arguments[0][0] == '#' {
		channelName := event.Arguments[0][1:len(event.Arguments[0])]
		channelid = discord.GetChannelByName(channelName)

		if len(channelid) == 0 {
			channelid = discord.CreateChannel(channelName, "IRC-Channels").ID
		}

		return
	}

	// DM
	channelid = discord.GetChannelByName(event.Nick)
	if len(channelid) == 0 {
		channelid = discord.CreateChannel(event.Nick, "IRC-DMs").ID
	}

	return
}

func onIrcConnected(_ *irc.Event) {
	channels, err := Dsession.GuildChannels(viper.GetString("discord.guild"))
	CheckError("Failed to fetch guild channels", err)

	for _, channel := range channels {

		// skip categories
		if len(channel.ParentID) == 0 {
			continue
		}

		//parent, err := dsession.Channel(channel.ParentID)
		parent, err := Dsession.State.Channel(channel.ParentID)
		CheckError("Failed to fetch parent channel", err)

		// only join channels, not DMs
		if parent.Name == "IRC-Channels" {
			IrcSession.Join("#" + channel.Name)
		}
	}
}

func onIrcMsg(event *irc.Event) {
	channelId := GetChannelId(event)
	discord.SendMessage(channelId, EnrichMsg(event.Message()), event.Nick)
}

// add mentions to IRC messages
func EnrichMsg(msg string) string {
	member, err := Dsession.State.Member(viper.GetString("discord.guild"), viper.GetString("discord.user"))
	CheckError("Failed to get client member", err)

	return strings.ReplaceAll(msg, IrcSession.GetNick(), member.Mention())
}