package main

import (
	"log"
	"time"

	"github.com/yanzay/tbot/v2"
)

// Regexp defining commands that could be sent to IRC bot
const commandRegexp string = `^@(\?|v\?|V\?|d\?|e\?|g\?|b\?|l(\?|e\?|t\?)|s\?|u(\?|\+\?))\s*\S+`

func timeStat(h tbot.UpdateHandler) tbot.UpdateHandler {
	return func(u *tbot.Update) {
		start := time.Now()
		h(u)
		log.Printf("User is: %d\t%s %s %s",
			u.Message.From.ID,
			u.Message.From.Username,
			u.Message.From.FirstName,
			u.Message.From.LastName)
		log.Printf("Handle time: %v", time.Now().Sub(start))
	}
}

// Handler for /start command
func (a *application) startHandler(m *tbot.Message) {
	// WELCOME_MESSAGE is defined in const.go
	a.Telegram.Client.SendMessage(m.Chat.ID, WELCOME_MESSAGE)
}

// Handler for commands related to IRC Pinobot
func (a *application) pinobotHandler(m *tbot.Message) {
	queryChannel <- botQuery{"Pinoclone", m}
}

// Handler for commands related to IRC Beholder bot
func (a *application) beholderHandler(m *tbot.Message) {
	queryChannel <- botQuery{"Beholder", m}
}

// Handler for !pom command and moon phase calculation
// Variable PoM of type pomRequest must be declared and init'd beforehand
func (a *application) pomHandler(m *tbot.Message) {
	// Save time of the request
	updateTime := time.Now()
	// If pom.jpg wasn't updated in an hour do an update
	if PoM.UpdatedAt.Hour()-updateTime.Hour() != 0 {
		err := PoM.updateImage()
		switch {
		// in case there was an error running xplanets send this error as a message
		case err != nil:
			PoM.Text = err.Error()
		// otherwise update pom.Text and save the update timestamp
		default:
			PoM.updateText()
			PoM.UpdatedAt = updateTime
		}
	}
	// Send the image back to Telegram with pom.Text as a caption
	app.Telegram.Client.SendPhotoFile(m.Chat.ID, "pom.jpg", tbot.OptCaption(PoM.Text))
}
