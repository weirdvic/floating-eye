package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/yanzay/tbot/v2"
)

// Handler for /start command
func (a *application) startHandler(m *tbot.Message) {
	// WELCOME_MESSAGE is defined in const.go
	a.Telegram.Client.SendMessage(m.Chat.ID, welcomeMessage)
}

// Handler for /announce command
func (a *application) announceHandler(m *tbot.Message) {
	// Only admin can make announces
	if m.Chat.ID == strconv.Itoa(a.Telegram.Admins[0]) {
		a.Telegram.Client.SendMessage(
			strconv.Itoa(a.Telegram.ForwardChat),
			strings.TrimPrefix(m.Text, "/announce "))
	} else {
		a.Telegram.Client.SendMessage(m.Chat.ID, "You can't handle my potions, traveller…")
	}
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
// Variable PoM of type pomRequest must be created at init
func (a *application) pomHandler(m *tbot.Message) {
	// Save time of the request
	updateTime := time.Now()
	// If pom.jpg wasn't updated in an hour do an update
	if PoM.UpdatedAt.Hour()-updateTime.Hour() != 0 {
		err := PoM.updateImage()
		/* in case there was an error running xplanets send this error as a message
		   otherwise update pom.Text and save the update timestamp */
		if err != nil {
			PoM.Text = err.Error()
		} else {
			PoM.updateText()
			PoM.UpdatedAt = updateTime
		}
	}
	// Send the image back to Telegram with pom.Text as a caption
	app.Telegram.Client.SendPhotoFile(m.Chat.ID, "pom.jpg", tbot.OptCaption(PoM.Text))
}

// Handler for !orcname command
func (a *application) orcnameHandler(m *tbot.Message) {
	a.Telegram.Client.SendMessage(m.Chat.ID, strings.Title(makeOrcName()))
}
