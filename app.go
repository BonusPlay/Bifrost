package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

var protocols map[uint]Protocol

func main() {
	protocols = make(map[uint]Protocol, 0)
	setupConfig()

	irc, err := IrcCreateBot()
	if err != nil {
		log.Fatal("Failed to startup IRC bot")
		return
	}
	log.Info("[IRC] started")
	protocols[IRC] = irc

	// init discord last as it will make others join channels on startup
	discord, err := DiscordCreateBot()
	if err != nil {
		log.Fatal("Failed to startup discord bot")
		return
	}
	log.Info("[Discord] started")
	protocols[Discord] = discord

	err = discord.SyncChannels()
	if err != nil {
		log.Fatal("Failed to sync channels")
		return
	}

	log.Info("Bridge is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	for _, protocol := range protocols {
		protocol.Close()
	}
}

func setupConfig() {
	viper.SetConfigFile("bifrost.cfg")
	viper.SetConfigType("json")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Failed to read config")
	}

	return
}

func SendMessage(msg TextMessage) {
	protocols[msg.ToProtocol].SendMessage(msg)
}

func JoinChannel(msg JoinChannelMessage) {
	protocols[msg.ToProtocol].JoinChannel(msg)
}

func LeaveChannel(msg LeaveChannelMessage) {
	protocols[msg.ToProtocol].LeaveChannel(msg)
}