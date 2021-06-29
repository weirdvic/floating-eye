package main

import (
	"fmt"
	"log"
	"time"

	"github.com/yanzay/tbot/v2"
)

// Regexp defining commands that could be sent to IRC bot
const commandRegexp string = `^@(\?|v\?|V\?|d\?|e\?|g\?|b\?|l(\?|e\?|t\?)|s\?|u(\?|\+\?))\s*\S+`

// botStat is used to log bot users
var botStat = make(map[int]string)

func timeStat(h tbot.UpdateHandler) tbot.UpdateHandler {
	return func(u *tbot.Update) {
		start := time.Now()
		h(u)
		log.Printf("Handle time: %v", time.Now().Sub(start))
	}
}

// Handler for /stat command
func (a *application) statHandler(m *tbot.Message) {
	logUser(botStat, m)
	if isAllowedAdmin(m.From.ID, a) {
		a.Telegram.Client.SendMessage(m.Chat.ID, fmt.Sprintf("Known users:\n%v", botStat))
		return
	}
	a.Telegram.Client.SendMessage(m.Chat.ID, "You are not allowed to use this command…")
}

// Handler for /start command
func (a *application) startHandler(m *tbot.Message) {
	logUser(botStat, m)
	// welcomeMessage is defined in const.go
	a.Telegram.Client.SendMessage(m.Chat.ID, welcomeMessage)
}

// Handler for commands related to IRC Pinobot
func (a *application) pinobotHandler(m *tbot.Message) {
	logUser(botStat, m)
	queryChannel <- botQuery{"Pinoclone", m}
}

// Handler for commands related to IRC Beholder bot
func (a *application) beholderHandler(m *tbot.Message) {
	logUser(botStat, m)
	queryChannel <- botQuery{"Beholder", m}
}

// Handler for !pom command and moon phase calculation
func (a *application) pomHandler(m *tbot.Message) {
	logUser(botStat, m)
	// Save time of the request
	updateTime := time.Now()
	// If pom.jpg wasn't updated in an hour do an update
	if pom.UpdatedAt.Hour()-updateTime.Hour() != 0 {
		err := updatePomImage()
		switch {
		// in case there was an error running xplanets send this error as a message
		case err != nil:
			pom.Text = err.Error()
		// otherwise update pom.Text and save the update timestamp
		default:
			pom.Text = getPomText()
			pom.UpdatedAt = updateTime
		}
	}
	// Send the image back to Telegram with pom.Text as a caption
	app.Telegram.Client.SendPhotoFile(m.Chat.ID, "pom.jpg", tbot.OptCaption(pom.Text))
}
