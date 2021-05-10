package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/yanzay/tbot/v2"
	"gopkg.in/irc.v3"
)

type Application struct {
	TgClient  *tbot.Client
	IrcClient *irc.Client
	Conf      AppConfig
}

type AppConfig struct {
	Tg  TgConfig
	Irc IrcConfig
}

type BotQuery struct {
	BotNick string
	Query   *tbot.Message
}

type TgConfig struct {
	Token string
}

type IrcConfig struct {
	Server string
	Port   string
	Nick   string
	Pass   string
	Name   string
	Bot    string
}

var (
	app             Application
	workers         sync.WaitGroup
	inboxChannel    = make(chan BotQuery, 100)
	responseChannel = make(chan string, 100)
)

func init() {
	// Load config from file config.json and decode it to tgconfig struct
	configfile, err := os.Open("config.json")
	if err != nil {
		log.Fatalln(err)
	}
	defer configfile.Close()
	decoder := json.NewDecoder(configfile)
	err = decoder.Decode(&app.Conf)
	if err != nil {
		log.Fatalln(err)
	} else {
		log.Printf("Config successfully loaded.")
	}
	// Create new Telegram bot with token from config
	tgBot := tbot.New(app.Conf.Tg.Token)
	log.Printf("Created new bot with token: %s", app.Conf.Tg.Token)
	app.TgClient = tgBot.Client()

	// Set middleware
	tgBot.Use(stat)

	// Set start or help message handler
	tgBot.HandleMessage("^/(start|help)$", app.startHandler)
	// Set main command handler
	tgBot.HandleMessage(commandRegexp, app.commandHandler)
	// Set !lastgame command handler
	tgBot.HandleMessage("^!(scores|sb|players|who|variant)\\s*$", app.beholderHandler)
	tgBot.HandleMessage("^!(whereis|streak|role|race)\\s*\\w*\\s*$", app.beholderHandler)
	tgBot.HandleMessage("^!(lastgame|asc|lastasc)\\s*\\w*\\s*\\w*$", app.beholderHandler)

	// Start the Telegram bot
	go func() {
		log.Println("Connecting to Telegram…")
		log.Fatalln(tgBot.Start())
	}()

	// Initialize IRC config
	config := irc.ClientConfig{
		Nick:    app.Conf.Irc.Nick,
		Pass:    app.Conf.Irc.Pass,
		User:    app.Conf.Irc.Nick,
		Name:    app.Conf.Irc.Name,
		Handler: ircHandlerFunc,
	}

	// Connect to IRC server
	conn, err := net.Dial("tcp", app.Conf.Irc.Server+":"+app.Conf.Irc.Port)
	if err != nil {
		log.Fatalln(err)
	}
	app.IrcClient = irc.NewClient(conn, config)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		app.IrcClient.Write("QUIT")
	}()
}

func inboxWorker(c <-chan BotQuery, a *Application) {
	for q := range c {
		a.IrcClient.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params:  []string{q.BotNick, q.Query.Text}})
		botResponse := <-responseChannel
		a.TgClient.SendMessage(q.Query.Chat.ID, botResponse)
	}
}

func main() {
	// Run IRC client
	go func() {
		log.Println("Connecting to IRC…")
		log.Fatalln(app.IrcClient.Run())
	}()

	workers.Add(1)
	go func() {
		log.Println("Starting inbox worker…")
		inboxWorker(inboxChannel, &app)
		workers.Done()
	}()

	workers.Wait()
}
