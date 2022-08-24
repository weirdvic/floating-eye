package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/yanzay/tbot/v2"
)

func timeStat(h tbot.UpdateHandler) tbot.UpdateHandler {
	return func(u *tbot.Update) {
		start := time.Now()
		h(u)
		log.Printf("User is: %d\t%s %s %s",
			u.Message.From.ID,
			u.Message.From.Username,
			u.Message.From.FirstName,
			u.Message.From.LastName)
		log.Printf("Handle time: %v", time.Since(start))
	}
}

// Handler for /start command
func (a *application) startHandler(m *tbot.Message) {
	// WELCOME_MESSAGE is defined in const.go
	a.Telegram.Client.SendMessage(m.Chat.ID, welcomeMessage)
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

func (a *application) pornHandler(m *tbot.Message) {
	// Check if "oglaf" directory exists
	_, err := os.Stat("oglaf")
	if os.IsNotExist(err) {
		log.Print("Oglaf images directory does not exist!")
		app.Telegram.Client.SendMessage(m.Chat.ID, "No porn for you!")
	}
	// Determine how many images in "oglaf" directory
	pattern := filepath.Join("oglaf", "*.png")
	pngList, err := filepath.Glob(pattern)
	if err != nil {
		log.Print("Files error:", err)
	}
	oglafPicNum := rand.Intn(len(pngList))
	app.Telegram.Client.SendPhotoFile(
		m.Chat.ID,
		filepath.Join("oglaf", fmt.Sprintf("%d.png", oglafPicNum)),
		tbot.OptCaption(fmt.Sprintf("Oglaf pic %d from %d", oglafPicNum, len(pngList))))
}
