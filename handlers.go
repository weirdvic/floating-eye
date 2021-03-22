package main

import (
	"log"
	"time"

	"github.com/yanzay/tbot/v2"
)

func stat(h tbot.UpdateHandler) tbot.UpdateHandler {
	return func(u *tbot.Update) {
		start := time.Now()
		h(u)
		log.Printf("Handle time: %v", time.Now().Sub(start))
	}
}

func (a *application) startHandler(m *tbot.Message) {
	a.client.SendMessage(m.Chat.ID, "This is a bot to query IRC knowlege base bot for NetHack.")
}
