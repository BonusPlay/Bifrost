package main

import (
	"crypto/tls"
	"github.com/spf13/viper"
	irc "github.com/thoj/go-ircevent"
)

type IrcProtocol struct {
	session *irc.Connection
}

var ready chan struct{}

func IrcCreateBot() (*IrcProtocol, error) {
	cert, err := tls.LoadX509KeyPair(viper.GetString("irc.cert"), viper.GetString("irc.key"))
	if err != nil {
		return nil, err
	}

	session := irc.IRC(viper.GetString("irc.nick"), viper.GetString("irc.user"))
	if viper.IsSet("irc.real_name") {
		session.RealName = viper.GetString("irc.real_name")
	}
	session.UseTLS = true
	session.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	err = session.Connect("chat.freenode.net:6697")
	if err != nil {
		return nil, err
	}

	session.AddCallback("001", onIrcConnected)
	session.AddCallback("PRIVMSG", onIrcMsg)
	//session.AddCallback("353", onNames) // NAMES command lists all users in channel
	//session.AddCallback("366", onNamesEnd) // NAMES end
	//IrcSession.AddCallback("671", onNames) // whois secure
	//IrcSession.AddCallback("*", onNames)

	ready = make(chan struct{})

	p := IrcProtocol{
		session: session,
	}
	go session.Loop()

	<- ready // wait for client to connect before returning
	return &p, nil
}

func (i IrcProtocol) SendMessage(msg TextMessage) {
	if msg.Channel.Public {
		i.session.Privmsg("#" + msg.Channel.Name, sanitizeIrcMsg(msg.Text))
	} else {
		i.session.Privmsg(msg.Channel.Name, sanitizeIrcMsg(msg.Text))
	}
}

func (i IrcProtocol) JoinChannel(msg JoinChannelMessage) {
	if msg.Channel.Public {
		i.session.Join("#" + msg.Channel.Name)
	}
}

func (i IrcProtocol) LeaveChannel(msg LeaveChannelMessage) {
	if msg.Channel.Public {
		i.session.Part("#" + msg.Channel.Name)
	}

}

func (i IrcProtocol) Close() {
	panic("implement me")
}

func onIrcConnected(event *irc.Event) {
	close(ready)
}

func onIrcMsg(event *irc.Event) {
	var channel Channel

	if event.Arguments[0][0] == '#' {
		channel = Channel{
			Name:   event.Arguments[0][1:len(event.Arguments[0])],
			Public: true,
		}
	} else {
		// DM
		channel = Channel{
			Name: event.Nick,
			Public: false,
		}
	}

	msg := TextMessage{
		FromProtocol: IRC,
		ToProtocol:   Discord,
		From:         event.Nick,
		Text:         event.Message(),
		Channel:      channel,
	}
	SendMessage(msg)
}

// replace discord emotes with text
func sanitizeIrcMsg(msg string) string {
	// cut to parts, etc

	return msg
}