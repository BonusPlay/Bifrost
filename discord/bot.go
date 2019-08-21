package discord

import (
	"fmt"
	. "github.com/BonusPlay/Bifrost/util"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"regexp"
	"strings"
)

var Commands map[string]Command

func SetupBot() {
	session, err := discordgo.New("Bot " + viper.GetString("discord.token"))
	Dsession = session
	CheckError("Discord bot failed to start", err)
	Dsession.StateEnabled = true

	Dsession.AddHandler(onDiscordMsg)
	Dsession.AddHandler(onChannelCreated)
	Dsession.AddHandler(onChannelDeleted)
	Dsession.AddHandler(onChannelEdited)

	Commands = make(map[string]Command)
	Commands["help"] = new(CHelp)
	Commands["names"] = CNames{
		Buf: map[string][]string{},
	}
	Commands["whois"] = new(CWhois)
}

func SendMessage(channelId string, msg string, username string) {

	webhooks, err := Dsession.ChannelWebhooks(channelId)
	CheckError("Failed to get channel webhooks", err)

	var webhookId string
	var webhookToken string

	for _, webhook := range webhooks {
		if webhook.Name == "Bifrost" {
			webhookId = webhook.ID
			webhookToken = webhook.Token
		}
	}

	// channel does not have a webhook setup
	if len(webhookId) == 0 {
		webhookId, webhookToken = SetupWebhook(channelId)
	}

	data := discordgo.WebhookParams{
		Username: username,
		Content:  msg,
	}

	err = Dsession.WebhookExecute(webhookId, webhookToken, false, &data)
	CheckError("Failed to execute webhook", err)
}

func SetupWebhook(channelId string) (webhookId string, token string) {
	webhook, err := Dsession.WebhookCreate(channelId, "Bifrost", "https://i.imgur.com/ul4i5RW.jpg")
	CheckError("Failed to create webhook", err)

	return webhook.ID, webhook.Token
}

// will return empty if channel was not found
func GetChannelByName(name string) (channelid string) {

	channels, err := Dsession.GuildChannels(viper.GetString("discord.guild"))
	CheckError("Failed to fetch guild channels", err)

	for _, channel := range channels {
		if channel.Name == name {
			return channel.ID
		}
	}

	// channel not found
	return ""
}

// parent - cateogy to put channel under
func CreateChannel(name string, parent string) (channel *discordgo.Channel) {

	channel, err := Dsession.GuildChannelCreate(viper.GetString("discord.guild"), name, discordgo.ChannelTypeGuildText)
	CheckError("Failed to create DM channel", err)

	data := &discordgo.ChannelEdit{
		ParentID: GetChannelByName(parent),
	}

	channel, err = Dsession.ChannelEditComplex(channel.ID, data)
	CheckError("Failed to edit DM channel", err)

	return
}

// replace discord emotes with text
func SanitizeMsg(msg *discordgo.MessageCreate) (ret string) {

	ret = msg.ContentWithMentionsReplaced()

	// finds discord encoded emotes
	r1 := regexp.MustCompile("<(:.*?:)([0-9]*?)>")

	// finds name of discord emotes
	r2 := regexp.MustCompile(":.*?:")

	ret = r1.ReplaceAllStringFunc(ret, func(s string) string {
		return r2.FindString(s)
	})

	return
}

func onChannelCreated(dsession *discordgo.Session, m *discordgo.ChannelCreate) {

	// skip categories
	if len(m.ParentID) == 0 {
		return
	}

	parent, err := Dsession.State.Channel(m.ParentID)
	CheckError("Failed to get parent channel", err)

	// we can ignore DMs here, since bot will auto-connect on first message
	if parent.Name == "IRC-Channels" {
		IrcSession.Join("#" + m.Name)
	}
}

func onChannelDeleted(dsession *discordgo.Session, m *discordgo.ChannelDelete) {

	// skip categories
	if len(m.ParentID) == 0 {
		return
	}

	parent, err := Dsession.State.Channel(m.ParentID)
	CheckError("Failed to get parent channel", err)

	// we can ignore DMs here, since bot will auto-connect on first message
	if parent.Name == "IRC-Channels" {
		IrcSession.Part("#" + m.Name)
	}
}

func onChannelEdited(dsession *discordgo.Session, m *discordgo.ChannelUpdate) {

	// skip categories
	if len(m.ParentID) == 0 {
		return
	}

	parent, err := Dsession.State.Channel(m.ParentID)
	CheckError("Failed to get parent channel", err)

	if parent.Name == "IRC-Channels" {
		_, err := dsession.ChannelMessageSend(m.ID, "Bifrost does not handle channel edits well")
		CheckError("Failed to send discord message", err)
	}
}

func onDiscordMsg(dsession *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by bots (includes IRC messages)
	if m.Author.Bot {
		return
	}

	// see if message is a command
	for _, mention := range m.Mentions {
		if mention.ID == Dsession.State.User.ID {
			parseCommand(m)
			return
		}
	}

	channel, err := Dsession.State.Channel(m.ChannelID)
	CheckError("Failed to fetch discord channel from message", err)
	parent, err := Dsession.State.Channel(channel.ParentID)
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

	IrcSession.Privmsg(channelName, SanitizeMsg(m))

	// get links to attachments and send them as well
	for _, attachment := range m.Attachments {
		IrcSession.Privmsg(channelName, fmt.Sprintf("%s: %s", attachment.Filename, attachment.URL))
	}
}

func parseCommand(msg *discordgo.MessageCreate) {
	// trim mention
	text := strings.Replace(msg.Content, "<@" + Dsession.State.User.ID + ">", "", -1)
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	parts := strings.Split(text, " ")

	if cmd, ok := Commands[parts[0]]; ok {
		cmd.Run(msg.Message)
	} else {
		_, err := Dsession.ChannelMessageSend(msg.ChannelID, "Unrecognized command")
		CheckError("Failed to send message", err)
	}
}