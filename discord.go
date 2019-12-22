package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"regexp"
	"strings"
)

type DiscordProtocol struct {
	session *discordgo.Session
}

func DiscordCreateBot() (*DiscordProtocol, error) {
	session, err := discordgo.New("Bot " + viper.GetString("discord.token"))
	if err != nil {
		return nil, err
	}

	session.StateEnabled = true

	session.AddHandler(onDiscordMsg)
	session.AddHandler(onChannelCreated)
	session.AddHandler(onChannelDeleted)
	session.AddHandler(onChannelEdited)

	p := DiscordProtocol{
		session: session,
	}
	err = p.session.Open()
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (p DiscordProtocol) SyncChannels() error {
	// join channels
	channels, err := p.session.GuildChannels(viper.GetString("discord.guild"))
	if err != nil {
		return err
	}

	for _, channel := range channels {
		// skip categories
		if len(channel.ParentID) == 0 {
			continue
		}

		parent, err := p.session.Channel(channel.ParentID)
		if err != nil {
			return err
		}

		msg := JoinChannelMessage{
			FromProtocol: Discord,
			Channel: Channel{
				Name:   channel.Name,
				Public: true,
			},
		}

		switch parent.Name {
		case "IRC-Channels":
			msg.ToProtocol = IRC
			break

		// only join channels, not DMs

		default:
			continue
		}

		JoinChannel(msg)
	}

	return nil
}

func (p DiscordProtocol) SendMessage(msg TextMessage) {
	channelId, err := p.channelIdByName(msg.Channel.Name)
	if err != nil {
		log.Error("[Discord] Failed to get channelID by name")
		return
	}

	if len(channelId) == 0 {
		msg := JoinChannelMessage{
			FromProtocol: msg.FromProtocol,
			ToProtocol:   msg.ToProtocol,
			Channel:      msg.Channel,
		}
		p.JoinChannel(msg)
	}

	webhooks, err := p.session.ChannelWebhooks(channelId)
	if err != nil {
		log.Error("[Discord] Failed to get webhooks by channelID")
	}

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
		webhookId, webhookToken, err = p.setupWebhook(channelId)
		if err != nil {
			log.Error("[Discord] Failed to setup webhook")
			return
		}
	}

	data := discordgo.WebhookParams{
		Username: msg.From,
		Content: msg.Text,
	}

	err = p.session.WebhookExecute(webhookId, webhookToken, false, &data)
	if err != nil {
		log.Error("[Discord] Failed to execute webhook")
	}
}

func (p DiscordProtocol) JoinChannel(msg JoinChannelMessage) {
	channel, err := p.session.GuildChannelCreate(viper.GetString("discord.guild"), msg.Channel.Name, discordgo.ChannelTypeGuildText)
	if err != nil {
		log.Error("[Discord] Failed to create guild channel")
		return
	}

	var parentName string

	switch msg.FromProtocol {
	case IRC:
		if msg.Channel.Public {
			parentName = "IRC-Channels"
		} else {
			parentName = "IRC-DMs"
		}
		break

	default:
		return
	}

	parent, err := p.channelIdByName(parentName)
	if err != nil {
		log.Error("[Discord] Failed to get channel by name")
		return
	}

	data := &discordgo.ChannelEdit{
		ParentID: parent,
	}

	channel, err = p.session.ChannelEditComplex(channel.ID, data)
	if err != nil {
		log.Error("[Discord] Failed to edit channel")
	}

	return
}

func (p DiscordProtocol) LeaveChannel(msg LeaveChannelMessage) {
	panic("implement me")
}

func (p DiscordProtocol) Close() {
	_ = p.session.Close()
}

// returns empty string (without error) if channel was not found
func (p DiscordProtocol) channelIdByName(name string) (string, error) {
	channels, err := p.session.GuildChannels(viper.GetString("discord.guild"))
	if err != nil {
		return "", err
	}

	for _, channel := range channels {
		// discord channels are only lower strings
		if strings.ToLower(channel.Name) == strings.ToLower(name) {
			return channel.ID, nil
		}
	}

	return "", nil
}

func (p DiscordProtocol) setupWebhook(channelID string) (string, string, error) {
	webhook, err := p.session.WebhookCreate(channelID, "Bifrost", "https://i.imgur.com/ul4i5RW.jpg")
	if err != nil {
		return "", "", err
	}

	return webhook.ID, webhook.Token, nil
}

// replace discord emotes with text
func sanitizeMsg(msg *discordgo.MessageCreate) (ret string) {

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

// add mentions to messages
func (p DiscordProtocol) enrichMsg(msg string) string {
	member, err := p.session.State.Member(viper.GetString("discord.guild"), viper.GetString("discord.user"))
	if err != nil {
		log.Error("[Discord] Failed to get client member")
		return msg
	}

	// TODO: get accurate IRC nickname
	return strings.ReplaceAll(msg, viper.GetString("irc.nick"), member.Mention())
}

func onChannelCreated(session *discordgo.Session, m *discordgo.ChannelCreate) {

	// skip categories
	if len(m.ParentID) == 0 {
		return
	}

	parent, err := session.State.Channel(m.ParentID)
	if err != nil {
		log.Error("[Discord] Failed to get parent channel")
		return
	}

	msg := JoinChannelMessage{
		FromProtocol: Discord,
	}

	switch parent.Name {
	case "IRC-Channels":
		msg.ToProtocol = IRC
		msg.Channel = Channel{
			Name:   m.Name,
			Public: true,
		}
		break

	// we can ignore "IRC-DMs", since IRC will auto join channel on 1st message

	default:
		return
	}

	JoinChannel(msg)
}

func onChannelDeleted(session *discordgo.Session, m *discordgo.ChannelDelete) {

	// skip categories
	if len(m.ParentID) == 0 {
		return
	}

	parent, err := session.State.Channel(m.ParentID)
	if err != nil {
		log.Error("[Discord] Failed to get parent channel")
		return
	}

	msg := LeaveChannelMessage{
		FromProtocol: Discord,
		Channel: Channel{
			Name:   m.Name,
		},
	}

	switch parent.Name {
	case "IRC-Channels":
		msg.ToProtocol = IRC
		msg.Channel.Public = true
		break

	case "IRC-DMs":
		msg.ToProtocol = IRC
		msg.Channel.Public = false
		break

	// ignore rest channels
	default:
		return
	}

	LeaveChannel(msg)
}

func onChannelEdited(session *discordgo.Session, m *discordgo.ChannelUpdate) {

	// skip categories
	if len(m.ParentID) == 0 {
		return
	}

	parent, err := session.State.Channel(m.ParentID)
	if err != nil {
		log.Error("[Discord] Failed to get parent channel")
		return
	}

	if parent.Name == "IRC-Channels" {
		// TODO: discord channel edit
		//_, err := dsession.ChannelMessageSend(m.ID, "Bifrost does not handle channel edits well")
		//CheckError("Failed to send discord message", err)
	}
}

func onDiscordMsg(session *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by bots (includes IRC messages)
	if m.Author.Bot {
		return
	}

	// see if message is a command
	for _, mention := range m.Mentions {
		if mention.ID == session.State.User.ID {
			// TODO: commands
			//parseCommand(m)
			return
		}
	}

	channel, err := session.State.Channel(m.ChannelID)
	if err != nil {
		log.Error("[Discord] Failed to get channel from message")
		return
	}
	parent, err := session.State.Channel(channel.ParentID)
	if err != nil {
		log.Error("[Discord] Failed to get parent channel from message")
		return
	}

	msg := TextMessage{
		FromProtocol: Discord,
		ToProtocol:   0,
		From:         viper.GetString("irc.nick"),
		Text:         sanitizeMsg(m),
		Channel:      Channel{
			Name: channel.Name,
		},
	}

	switch parent.Name {
	case "IRC-DMs":
		msg.Channel.Public = false
		msg.ToProtocol = IRC
		break

	case "IRC-Channels":
		msg.Channel.Public = true
		msg.ToProtocol = IRC
		break

	// ignore rest channels
	default:
		return
	}

	SendMessage(msg)

	// get links to attachments and send them as well
	for _, attachment := range m.Attachments {
		msg.Text = fmt.Sprintf("attachment %s: %s", attachment.Filename, attachment.URL)
		SendMessage(msg)
	}
}