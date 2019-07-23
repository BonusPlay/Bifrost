package discord

import (
	. "github.com/BonusPlay/Bifrost/util"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"regexp"
)

func SetupBot() (session *discordgo.Session) {
	session, err := discordgo.New("Bot " + viper.GetString("discord.token"))
	CheckError("Discord bot failed to start", err)

	session.StateEnabled = true

	return
}

// will return empty if channel was not found
func GetChannelByName(session *discordgo.Session, name string) (channelid string) {

	channels, err := session.GuildChannels(viper.GetString("discord.guild"))
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
func CreateDMChannel(session *discordgo.Session, name string, parent string) (channel *discordgo.Channel) {

	channel, err := session.GuildChannelCreate(viper.GetString("discord.guild"), name, discordgo.ChannelTypeGuildText)
	CheckError("Failed to create DM channel", err)

	data := &discordgo.ChannelEdit{
		ParentID: GetChannelByName(session, parent),
	}

	channel, err = session.ChannelEditComplex(channel.ID, data)
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
