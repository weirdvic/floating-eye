package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/yanzay/tbot/v2"
	"gopkg.in/irc.v3"
)

type application struct {
	Telegram struct {
		Client *tbot.Client
		Token  string `json:"token"`
		Admins []int  `json:"admins"`
	} `json:"telegram"`
	IRC struct {
		Client *irc.Client
		Server string   `json:"server"`
		Port   int      `json:"port"`
		Nick   string   `json:"nick"`
		Pass   string   `json:"pass"`
		Name   string   `json:"name"`
		Bots   []string `json:"bots"`
	} `json:"irc"`
}

type botQuery struct {
	BotNick string
	Query   *tbot.Message
}

var (
	//go:embed config.json
	configFile []byte
)

// askBot is a simple wrapper to send message to the IRC bot
func askBot(nick, text string) {
	app.IRC.Client.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params:  []string{nick, text}})
}

// getMonsterName parses string to extract monster's name
func getMonsterName(r *regexp.Regexp, s string) (name string, e error) {
	if !r.MatchString(s) {
		return "", errors.New("provided string does not contain a monster name")
	}
	match := r.FindStringSubmatch(s)
	if match[2] == "" {
		name = match[1]
		return name, nil
	}
	// else
	name = match[2]
	return name, nil
}

// checkBotName checks if bot name is in allowed list
func (a *application) checkBotName(item string) bool {
	for _, v := range a.IRC.Bots {
		if item == v {
			return true
		}
	}
	return false
}

// queryWorker reads from inboxChannel, passes the query text to IRC,
// awaits for response from bot and sends the response text back to Telegram
func queryWorker(c <-chan botQuery) {
	// This regexp is used to filter IRC color codes from Pinoclone's response
	colorFilter := regexp.MustCompile(`\(.*\d{1,2},\d{1,2}(\S).*\)|\[\s+\d{1,2}(\S+)\s+\]`)
	monsterNameFilter := regexp.MustCompile(`^([\w+\s+-]+)\s\[|~\d+~\s([\w+\s+-]+)\s\[`)
	for q := range c {
		askBot(q.BotNick, q.Query.Text)
		// Read response from the channel
		botResponse := <-responseChannel
		// Filter IRC color codes and replace parentheses to brackets
		botResponse = colorFilter.ReplaceAllString(botResponse, "[ $1$2 ]")
		// Split response to lines by '|' symbol
		botResponse = strings.ReplaceAll(botResponse, "|", "\n")
		// In case we're working on monster query
		if q.BotNick == "Pinoclone" {
			// Parsing monster's name
			monsterName, err := getMonsterName(monsterNameFilter, botResponse)
			if err == nil {
				fileName := fmt.Sprintf("monsters/%s.png", strings.ToUpper(monsterName))
				// If monster's image is not available, set placeholder image
				if _, err = os.Stat(fileName); err != nil {
					log.Println("Image not found: ", fileName)
					fileName = "monsters/WARNING 0.png"
				}
				// Send image with caption
				app.Telegram.Client.SendPhotoFile(
					q.Query.Chat.ID,
					fileName,
					tbot.OptCaption(botResponse),
					tbot.OptReplyToMessageID(q.Query.MessageID),
				)
			} else {
				log.Println(err)
				app.Telegram.Client.SendMessage(q.Query.Chat.ID, botResponse)
			}
		} else {
			// Send just text for other queries
			app.Telegram.Client.SendMessage(q.Query.Chat.ID, botResponse)
		}
	}
}

// readConfig decodes embedded config.json file to struct
func (a *application) init() {
	// Decode embedded config file to app struct
	err := json.Unmarshal(configFile, a)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Config successfully loaded.")
	log.Print("Available bots are: ")
	log.Println(a.IRC.Bots)
	// Check if dependency commands are available in the system
	var commands = []string{"pom", "xplanet"}
	for _, v := range commands {
		path, err := exec.LookPath(v)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Command %s is located at %s\n", v, path)
	}
	// Check if "monsters" directory exists
	_, err = os.Stat("monsters")
	if os.IsNotExist(err) {
		log.Fatal("Monster images directory does not exist!")
	}

	log.Println("All checks passed…")
}

// Send message to admin on shutdown
func (a *application) shutdown(reason string) {
	a.Telegram.Client.SendMessage(
		strconv.Itoa(a.Telegram.Admins[0]),
		fmt.Sprintf("Shutting down on: %s", reason))
	a.IRC.Client.Write("QUIT")
}
