package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/yanzay/tbot/v2"
)

type botConfig struct {
	Token string
}

var (
	app       application
	bot       tbot.Server
	botconfig botConfig
	token     string
)

type application struct {
	client *tbot.Client
}

func init() {
	// Load config from file config.json and decode it to botConfig struct
	configfile, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer configfile.Close()
	decoder := json.NewDecoder(configfile)
	err = decoder.Decode(&botconfig)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Create new bot with token from config
	bot := tbot.New(botconfig.Token)
	log.Printf("Created new bot with token: %s", botconfig.Token)
	app.client = bot.Client()
	bot.Use(stat)
	// Define message handlers
	bot.HandleMessage("/start", app.startHandler)
	log.Fatal(bot.Start())
}
