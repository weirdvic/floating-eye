package main

import (
	"log"
	"time"

	"github.com/yanzay/tbot/v2"
	"gopkg.in/irc.v3"
)

// Regexp defining commands that could be sent to IRC bot
// const commandRegexp string = "^!(help|commands|ping|time|tell|hello|lastgame|asc|lastasc|scores|setmintc|whereis|players|who|streak|role|race|variant|rng|pom|(\\d{1,2}d\\d{1,4}))\\s*"
const commandRegexp string = "^@(\\?|v\\?|V\\?|d\\?|e\\?|g\\?|b\\?|l(\\?|e\\?|t\\?)|s\\?|u(\\?|\\+\\?))\\s*\\S+"

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

Available monster query commands are:
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

Other commands are:
!lastgame [variant] [player] – display link to dumplog of last game ended.
!asc [player] [variant] – listing of all ascended games for a particular player (all variants or specified).
!lastasc [variant] [player] – dumplog for last ascended game.
!whereis [player] – shows variant and location within the game of the specified player.
!streak [player] – shows how many games a player has won in a row without dying.
!role [variant] – suggest a role for specified variant, or a variant and role.
!race [variant] – as above, for race.
!scores or !sb – provides you with a link to the Hardfought scoreboard of all variants hosted.
!players or !who – displays a list of all players currently online and which variant they are playing.

Where commands take the name of a variant, the following aliases are accepted:
nh343:  nh343 nethack 343
nh363:  nh363 363 363-hdf
nh370:  nh370 370 370-hdf
nh13d:  nh13d 13d
gh:  grunt grunthack
un:  unnethack unh
fh:  fiqhack
4k:  nhfourk nhf fourk
dnh:  dnethack dn
dyn:  dynahack dyn
nh4:  nethack4 n4
sp:  sporkhack spork
slex:  slex
xnh:  xnethack xnh
spl:  splicehack spl
slshm:  slashem slshm
ndnh:  notdnethack ndnh
evil:  evilhack evil
slth:  slashthem slth
`)
}

func (a *Application) commandHandler(m *tbot.Message) {
	inboxChannel <- BotQuery{app.Conf.Irc.Bot, m}
}

func (a *Application) beholderHandler(m *tbot.Message) {
	inboxChannel <- BotQuery{"Beholder", m}
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
	case m.Command == "PRIVMSG" && (m.Name == app.Conf.Irc.Bot || m.Name == "Beholder"):
		responseChannel <- m.Trailing()
	default:
		log.Println(m.Command, m.Params)
	}
},
)
