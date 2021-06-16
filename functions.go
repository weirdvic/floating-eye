package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/yanzay/tbot/v2"
	"gopkg.in/irc.v3"
)

// askBot is a simple wrapper to send message to the IRC bot
func askBot(nick, text string) {
	app.IRC.Client.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params:  []string{nick, text}})
}

// queryWorker reads from inboxChannel, passes the query text to IRC,
// awaits for response from bot and sends the response text back to Telegram
func queryWorker(c <-chan BotQuery) {
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
				fileName := fmt.Sprintf("mon/%s.png", strings.ToUpper(monsterName))
				// If monster's image is not available, set placeholder image
				if _, err = os.Stat(fileName); err != nil {
					log.Println("Image not found: ", fileName)
					fileName = "mon/WARNING 0.png"
				}
				// Send image with caption
				app.Telegram.Client.SendPhotoFile(q.Query.Chat.ID, fileName, tbot.OptCaption(botResponse))
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

// getMonsterName parses string to extract monster's name
func getMonsterName(r *regexp.Regexp, s string) (name string, e error) {
	if r.MatchString(s) != true {
		return "", errors.New("Provided string does not contain a monster name!")
	}
	match := r.FindStringSubmatch(s)
	if match[2] == "" {
		name = match[1]
		return name, nil
	} else {
		name = match[2]
		return name, nil
	}
}

// isAllowedBot checks if bot name is in allowed list
func isAllowedBot(item string, a *Application) bool {
	for _, s := range a.IRC.Bots {
		if item == s {
			return true
		}
	}
	return false
}

// isAllowedAdmin checks if bot name is in allowed list
func isAllowedAdmin(id int, a *Application) bool {
	for _, s := range a.Telegram.Admins {
		if id == s {
			return true
		}
	}
	return false
}

// logUser checks if user interacted with bot before and logs user's command
func logUser(stat map[int]string, m *tbot.Message) {
	if _, ok := stat[m.From.ID]; !ok {
		stat[m.From.ID] = m.From.Username
	}
	log.Printf("Incoming message from: ID: %d Name: %s", m.From.ID, m.From.Username)
	log.Printf("Command: %s", m.Text)
}

// checkError is a simple wrapper for "if err != nil" construction
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
