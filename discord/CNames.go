package discord

import (
	"fmt"
	. "github.com/BonusPlay/Bifrost/util"
	"github.com/bwmarrin/discordgo"
	"sort"
)

type CNames struct {
	// this is used because IRC is retarded and sends message in parts
	// key is channel name, [] is array of names
	Buf map[string][]string
}

func (_ CNames) Description() string {
	return "list all users in channel"
}

func (_ CNames) Run(msg *discordgo.Message) {
	channel, err := Dsession.State.Channel(msg.ChannelID)
	CheckError("Failed to fetch channel", err)

	// TODO: escape markdown
	// TODO: buffer names?
	IrcSession.SendRawf("NAMES #%s freenode", channel.Name)
}

func (cmd *CNames) AtEnd(channelName string) {
	names := cmd.Buf[channelName]
	sort.Strings(names)
	delete(cmd.Buf, channelName)

	channelId := GetChannelByName(channelName)

	var longest [3]int

	for i := 0; i < 3; i++ {
		longest[i] = 0
		for _, val := range names[:len(names) / 3] {
			if len(val) > longest[i] {
				longest[i] = len(val)
			}
		}
	}

	// calculate height of a single message
	sum := 0
	for i := range longest {
		sum += i
	}
	h := 2000 / sum

	// divides into parts
	var parts [][]string
	for i := 0; i < len(names); i += h {
		end := i + h

		if end > len(names) {
			end = len(names)
		}

		parts = append(parts, names[i:end])
	}

	for i := 0; i < len(parts) / 3; i++ {
		result := "```\n"

		for k := 0; k < h; k++ {
			for j := 0; j < 3; j++ {
				result += fmt.Sprintf("%*s", longest[j], parts[i * j][k])
			}
			result += "\n"
		}
		result += "```"
		_, err := Dsession.ChannelMessageSend(channelId, result)
		CheckError("Failed to send discord message", err)
	}
}