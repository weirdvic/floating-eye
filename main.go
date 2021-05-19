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

type TgConfig struct {
	Token string
}

type IrcConfig struct {
	Server string
	Port   string
	Nick   string
	Pass   string
	Name   string
	Bots   []string
}

type BotQuery struct {
	BotNick string
	Query   *tbot.Message
}

var (
	app             Application
	workers         sync.WaitGroup
	queryChannel    = make(chan BotQuery, 100)
	responseChannel = make(chan string, 100)
)

func init() {
	// Load config from file config.json and decode it to tgconfig struct
	configfile, err := os.Open("config.json")
	checkError(err)
	defer configfile.Close()
	decoder := json.NewDecoder(configfile)
	err = decoder.Decode(&app.Conf)
	checkError(err)
	log.Println("Config successfully loaded.")
	log.Print("Available bots are: ")
	log.Println(app.Conf.Irc.Bots)

	// Create new Telegram bot with token from config
	tgBot := tbot.New(app.Conf.Tg.Token)
	log.Printf("Created new bot with token: %s", app.Conf.Tg.Token)
	app.TgClient = tgBot.Client()

	// Set middleware
	tgBot.Use(stat)

	// Set start or help message handler
	tgBot.HandleMessage("^/(start|help)$", app.startHandler)
	// Set Pinobot IRC bot handlers
	tgBot.HandleMessage(commandRegexp, app.pinobotHandler)
	// Set Beholder IRC bot handlers
	tgBot.HandleMessage("^!(scores|sb|players|who|variant)\\s*$", app.beholderHandler)
	tgBot.HandleMessage("^!(whereis|streak|role|race)\\s*\\w*\\s*$", app.beholderHandler)
	tgBot.HandleMessage("^!(lastgame|asc|lastasc)\\s*\\w*\\s*\\w*$", app.beholderHandler)
	// Set !pom command handler
	tgBot.HandleMessage("^!pom\\.*", app.pomHandler)

	// Start the Telegram bot
	go func() {
		log.Println("Connecting to Telegram…")
		err := tgBot.Start()
		checkError(err)
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
	checkError(err)
	app.IrcClient = irc.NewClient(conn, config)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		app.IrcClient.Write("QUIT")
	}()
}

// askBot is a simple wrapper to send message to the IRC bot
func askBot(nick, text string) {
	app.IrcClient.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params:  []string{nick, text}})
}

// queryWorker reads from inboxChannel, passes the query text to IRC,
// awaits for response from bot and sends the response text back to Telegram
func queryWorker(c <-chan BotQuery) {
	for q := range c {
		askBot(q.BotNick, q.Query.Text)
		botResponse := <-responseChannel
		app.TgClient.SendMessage(q.Query.Chat.ID, botResponse)
	}
}

// goodSender checks if sender is in allowed senders list
func goodSender(sender string) bool {
	for _, s := range app.Conf.Irc.Bots {
		if sender == s {
			return true
		}
	}
	return false
}

// checkError is a simple wrapper for "if err != nil" construction
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Run IRC client
	go func() {
		log.Println("Connecting to IRC…")
		err := app.IrcClient.Run()
		checkError(err)
	}()

	// Run main worker and wait
	workers.Add(1)
	go func() {
		log.Println("Starting inbox worker…")
		queryWorker(queryChannel)
		workers.Done()
	}()

	workers.Wait()
}
