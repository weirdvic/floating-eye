package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/yanzay/tbot/v2"
	"gopkg.in/irc.v3"
)

type Application struct {
	Telegram struct {
		Client *tbot.Client
		Token  string `json:token`
		Admins []int  `json:admins`
	} `json:telegram`
	IRC struct {
		Client *irc.Client
		Server string   `json:server`
		Port   int      `json:port`
		Nick   string   `json:nick`
		Pass   string   `json:pass`
		Name   string   `json:name`
		Bots   []string `json:bots`
	} `json:irc`
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
	pom             pomRequest
	botStat         = make(map[int]string)
)

func init() {
	// Load config from file config.json and decode it to tgconfig struct
	configfile, err := os.Open("config.json")
	checkError(err)
	defer configfile.Close()
	decoder := json.NewDecoder(configfile)
	err = decoder.Decode(&app)
	checkError(err)
	log.Println("Config successfully loaded.")
	log.Print("Available bots are: ")
	log.Println(app.IRC.Bots)

	// Initialize Phase of Moon structure and update pom.jpg
	pom.Updated = time.Now()
	pom.Text = getPomText()
	pom.ImageArgs = []string{"-origin", "earth", "-body", "moon", "-num_times", "1", "-output", "pom.jpg", "-geometry", "300x300"}
	err = updatePomImage(pom.ImageArgs)
	checkError(err)

	// Create new Telegram bot with token from config
	tgBot := tbot.New(app.Telegram.Token)
	log.Printf("Created new bot with token: %s", app.Telegram.Token)
	app.Telegram.Client = tgBot.Client()

	// Set middleware
	tgBot.Use(stat)

	// Set start or help message handler
	tgBot.HandleMessage(`^/(start|help)$`, app.startHandler)
	// Set stat message handler
	tgBot.HandleMessage(`^/stat$`, app.statHandler)
	// Set Pinobot IRC bot handlers
	tgBot.HandleMessage(commandRegexp, app.pinobotHandler)
	// Set Beholder IRC bot handlers
	tgBot.HandleMessage(`^!(scores|sb|players|who|variant)\s*$`, app.beholderHandler)
	tgBot.HandleMessage(`^!(whereis|streak|role|race)\s*\w*\s*$`, app.beholderHandler)
	tgBot.HandleMessage(`^!(lastgame|asc|lastasc)\s*\w*\s*\w*$`, app.beholderHandler)
	// Set !pom command handler
	tgBot.HandleMessage(`^!pom\.*`, app.pomHandler)

	// Start the Telegram bot
	go func() {
		log.Println("Connecting to Telegram…")
		err := tgBot.Start()
		checkError(err)
	}()

	// Initialize IRC config
	config := irc.ClientConfig{
		Nick:    app.IRC.Nick,
		Pass:    app.IRC.Pass,
		User:    app.IRC.Nick,
		Name:    app.IRC.Name,
		Handler: ircHandlerFunc,
	}

	// Connect to IRC server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", app.IRC.Server, app.IRC.Port))
	checkError(err)
	app.IRC.Client = irc.NewClient(conn, config)

	// QUIT from IRC on SIGTERM
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		app.IRC.Client.Write("QUIT")
	}()
}

func main() {
	// Run IRC client
	go func() {
		log.Println("Connecting to IRC…")
		err := app.IRC.Client.Run()
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
