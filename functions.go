package main

import (
	"github.com/yanzay/tbot/v2"
	"gopkg.in/irc.v3"
)

func inboxWorker(c <-chan *tbot.Message, a *Application) {
	for m := range c {
		queryChannel <- m.Text
		botResponse := <-responseChannel
		m.Text = botResponse
		outboxChannel <- m
	}
}

func queryWorker(c <-chan string, a *Application) {
	for m := range c {
		a.IrcClient.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params:  []string{app.Conf.Irc.Bot, m},
		},
		)
	}
}

func outboxWorker(c <-chan *tbot.Message, a *Application) {
	for m := range c {
		a.TgClient.SendMessage(m.Chat.ID, m.Text)
	}
}
