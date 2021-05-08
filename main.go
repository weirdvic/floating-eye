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
	Server  string
	Port    string
	Nick    string
	Pass    string
	Name    string
	Channel string
	Bot     string
}

var (
	app Application
	wg  sync.WaitGroup
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
}

func main() {
	// Create new Telegram bot with token from config
	tgBot := tbot.New(app.Conf.Tg.Token)
	log.Printf("Created new bot with token: %s", app.Conf.Tg.Token)
	app.TgClient = tgBot.Client()
	// Set middleware
	tgBot.Use(stat)
	// Set message handlers
	tgBot.HandleMessage("/start", app.startHandler)
	tgBot.HandleMessage("/test\\s+\\S", app.testHandler)

	// Start the Telegram bot
	wg.Add(1)
	go func() {
		log.Println("Connecting to Telegram…")
		log.Fatalln(tgBot.Start())
		wg.Done()
		log.Println("Telegram disconnected!")
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
	defer conn.Close()

	// Create and run IRC client
	wg.Add(1)
	go func() {
		app.IrcClient = irc.NewClient(conn, config)
		log.Println("Connecting to IRC…")
		log.Fatalln(app.IrcClient.Run())
		wg.Done()
		log.Println("IRC disconnected!")
	}()

	// Send QUIT on SIGTERM
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("Shutting down…")
		err := app.IrcClient.Write("QUIT")
		if err != nil {
			log.Fatalln(err)
		}
	}()

	// Wait for goroutines to finish
	wg.Wait()
}
