package main

import (
	"github.com/BonusPlay/Bifrost/discord"
	ircbot "github.com/BonusPlay/Bifrost/irc"
	. "github.com/BonusPlay/Bifrost/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	err := setupConfig()
	CheckError("Failed to load config\n", err)

	ircbot.SetupBot()
	discord.SetupBot()

	err = Dsession.Open()
	// h
	time.Sleep(3 * time.Second)
	CheckError("Failed to start discord bot", err)
	go IrcSession.Loop()

	log.Info("Bridge is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	IrcSession.Quit()
	_ = Dsession.Close()
}

func setupConfig() (err error) {
	viper.SetConfigFile("bifrost.cfg")
	viper.SetConfigType("json")

	err = viper.ReadInConfig()
	CheckError("Failed to read config", err)

	return
}
