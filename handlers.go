package main

import (
	"log"
	"strings"
	"time"

	"github.com/yanzay/tbot/v2"
	"gopkg.in/irc.v3"
)

func stat(h tbot.UpdateHandler) tbot.UpdateHandler {
	return func(u *tbot.Update) {
		start := time.Now()
		h(u)
		log.Printf("Handle time: %v", time.Now().Sub(start))
	}
}

func (a *Application) startHandler(m *tbot.Message) {
	a.TgClient.SendMessage(m.Chat.ID, "This is a bot to query IRC knowlege base bot for NetHack.")
}

func (a *Application) testHandler(m *tbot.Message) {
	log.Println("Sending to " + a.Conf.Irc.Bot)
	a.IrcClient.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params:  []string{a.Conf.Irc.Bot, strings.TrimPrefix(m.Text, "/test ")},
	},
	)
}

var ircHandlerFunc = irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
	switch {
	// Handle WELCOME event
	case m.Command == "001":
		c.Write("JOIN " + app.Conf.Irc.Channel)
		log.Println("Joined " + app.Conf.Irc.Channel)
	// Handle PING command
	case m.Command == "PING":
		log.Println("PING received…")
		c.Write("PONG")
		log.Println("PONG was sent…")
	// Handle messages from channel
	case m.Command == "PRIVMSG" && m.Params[0] == app.Conf.Irc.Channel:
		log.Println("Message from channel" + app.Conf.Irc.Channel)
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params:  []string{m.Params[0], m.Trailing()},
		},
		)
	// Handle private messages
	case m.Command == "PRIVMSG" && m.Name == app.Conf.Irc.Bot:
		log.Println("Message from user " + m.Name)
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params:  []string{app.Conf.Irc.Bot, m.Trailing()},
		},
		)
	default:
		log.Println(m.Command, m.Params)
	}
},
)
