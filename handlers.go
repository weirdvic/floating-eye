package main

import (
	"log"
	"time"

	"github.com/yanzay/tbot/v2"
	"gopkg.in/irc.v3"
)

// Regexp defining commands that could be sent to IRC bot
// const commandRegexp string = "^!(help|commands|ping|time|tell|hello|lastgame|asc|lastasc|scores|setmintc|whereis|players|who|streak|role|race|variant|rng|pom|(\\d{1,2}d\\d{1,4}))\\s*"
const commandRegexp string = "^@(\\?|v\\?|V\\?|d\\?|e\\?|g\\?|b\\?|l(\\?|e\\?|t\\?)|s\\?|u(\\?|\\+\\?))\\s*.+"

var (
	inboxChannel    = make(chan *tbot.Message, 100)
	queryChannel    = make(chan string, 100)
	responseChannel = make(chan string, 100)
	outboxChannel   = make(chan *tbot.Message, 100)
)

func stat(h tbot.UpdateHandler) tbot.UpdateHandler {
	return func(u *tbot.Update) {
		start := time.Now()
		h(u)
		log.Printf("Handle time: %v", time.Now().Sub(start))
	}
}

func (a *Application) startHandler(m *tbot.Message) {
	a.TgClient.SendMessage(m.Chat.ID,
		`This is a bot to query NetHack monsters stats.
Available commands are:
@?[monster] or @v?[monster]:  NetHack 3.7
@V?[monster]:  NetHack 3.4.3 / 3.6.x
@d?[monster]:  dNetHack
@e?[monster]:  EvilHack
@g?[monster]:  GruntHack
@b?[monster]:  NetHack Brass
@l?[monster]:  Slash’EM
@le?[monster]:  Slash’EM Extended
@lt?[monster]:  SlashTHEM
@s?[monster]:  SporkHack
@u?[monster]:  UnNetHack
@u+?[monster]:  UnNetHackPlus
`)
}

func (a *Application) commandHandler(m *tbot.Message) {
	inboxChannel <- m
}

var ircHandlerFunc = irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
	switch {
	// Handle WELCOME event
	case m.Command == "001":
		// Identify to the NickServ
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params:  []string{"NickServ", app.Conf.Irc.Nick, app.Conf.Irc.Pass},
		})
	// Handle PING command
	case m.Command == "PING":
		c.Write("PONG")
	// Handle private messages
	case m.Command == "PRIVMSG" && m.Name == app.Conf.Irc.Bot:
		responseChannel <- m.Trailing()
	default:
		log.Println(m.Command, m.Params)
	}
},
)
