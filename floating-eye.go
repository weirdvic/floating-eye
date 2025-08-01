package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/yanzay/tbot/v2"
	"gopkg.in/irc.v4"
)

var (
	app             application
	workers         sync.WaitGroup
	queryChannel    = make(chan botQuery) //
	responseChannel = make(chan string)
)

func init() {
	app.init()
	PoM.init()

	// Create new Telegram bot with token from config
	tgBot := tbot.New(app.Telegram.Token)
	tgBot.Use(stat)
	log.Printf("Created new bot…")
	app.Telegram.Client = tgBot.Client()

	// Set start or help message handler
	tgBot.HandleMessage(`^/(start|help)$`, app.startHandler)
	// Set announce handler
	tgBot.HandleMessage(`^/announce\s\S+`, app.announceHandler)
	// Set Pinobot IRC bot handlers
	tgBot.HandleMessage(commandRegexp, app.pinobotHandler)
	// Set Beholder IRC bot handlers
	tgBot.HandleMessage(
		`^!(scores|sb|players|who|variant)\s*$`, app.beholderHandler)
	tgBot.HandleMessage(
		`^!(whereis|streak|role|race)\s*\w*\s*$`, app.beholderHandler)
	tgBot.HandleMessage(
		`^!(lastgame|asc|lastasc)\s*\w*\s*\w*$`, app.beholderHandler)
	// Set !pom command handler
	tgBot.HandleMessage(`(^/|^!)pom\.*`, app.pomHandler)
	// Set !orcname command handler
	tgBot.HandleMessage(`(^/|^!)orcname\.*`, app.orcnameHandler)
	// Set !quaff command handler
	tgBot.HandleMessage(`(^/|^!)(quaff|drink|živeli)\.*`, app.quaffHandler)

	// Start the Telegram bot
	log.Println("Connecting to Telegram…")
	go func(bot *tbot.Server) {
		err := bot.Start()
		if err != nil {
			log.Fatal(err)
		}
	}(tgBot)

	// Initialize IRC config
	config := irc.ClientConfig{
		Nick: app.IRC.Nick,
		Pass: app.IRC.Pass,
		User: app.IRC.Nick,
		Name: app.IRC.Name,
		Handler: irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
			switch {
			// Handle WELCOME event
			case m.Command == "001":
				c.Writef("MODE %v -R", app.IRC.Nick)
				// Identify to the NickServ
				c.WriteMessage(&irc.Message{
					Command: "PRIVMSG",
					Params:  []string{"NickServ", app.IRC.Nick, app.IRC.Pass},
				})
				// Join channels
				c.Write("JOIN #hardfought,#tnnt")
			// Handle PING command
			case m.Command == "PING":
				c.Write("PONG")
			// Write private messages from trusted senders to the responseChannel to be picked up by queryWorker
			case m.Command == "PRIVMSG" && app.checkBotName(m.Name) && !c.FromChannel(m):
				responseChannel <- m.Trailing()
			case m.Command == "PRIVMSG" && c.FromChannel(m) && (m.Name == "Beholder" || m.Name == "Croesus"):
				app.parseChatMessage(m.Trailing())
			default:
				log.Println(m.Command, m.Params)
			}
		},
		),
	}

	// Connect to IRC server
	conn, err := net.Dial("tcp", net.JoinHostPort(app.IRC.Server, strconv.Itoa(app.IRC.Port)))
	if err != nil {
		log.Fatal(err)
	}
	app.IRC.Client = irc.NewClient(conn, config)

	// Send /QUIT to IRC on SIGTERM
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(a *application) {
		<-c
		a.shutdown("SIGTERM")
	}(&app)
}

func main() {
	// Start IRC client
	log.Println("Connecting to IRC…")
	go func(c *irc.Client) {
		err := c.Run()
		if err != nil {
			log.Fatal(err)
		}
	}(app.IRC.Client)

	// Start main worker and wait
	log.Println("Starting inbox worker…")
	workers.Add(1)
	go func(c chan botQuery) {
		queryWorker(c)
		workers.Done()
	}(queryChannel)

	workers.Wait()
}
