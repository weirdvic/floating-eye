package main

import (
	"fmt"
	"log"
	"time"

	"github.com/yanzay/tbot/v2"
	"gopkg.in/irc.v3"
)

// Regexp defining commands that could be sent to IRC bot
const commandRegexp string = `^@(\?|v\?|V\?|d\?|e\?|g\?|b\?|l(\?|e\?|t\?)|s\?|u(\?|\+\?))\s*\S+`

func stat(h tbot.UpdateHandler) tbot.UpdateHandler {
	return func(u *tbot.Update) {
		start := time.Now()
		h(u)
		log.Printf("Handle time: %v", time.Now().Sub(start))
	}
}

func (a *Application) statHandler(m *tbot.Message) {
	logUser(botStat, m)
	if isAllowedAdmin(m.From.ID, a) {
		a.Telegram.Client.SendMessage(m.Chat.ID, fmt.Sprintf("Known users:\n%v", botStat))
		return
	}
	a.Telegram.Client.SendMessage(m.Chat.ID, "You are not allowed to use this command…")
}

func (a *Application) startHandler(m *tbot.Message) {
	logUser(botStat, m)
	a.Telegram.Client.SendMessage(m.Chat.ID,
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
!pom - display current phase of moon.

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

func (a *Application) pinobotHandler(m *tbot.Message) {
	logUser(botStat, m)
	queryChannel <- BotQuery{"Pinoclone", m}
}

func (a *Application) beholderHandler(m *tbot.Message) {
	logUser(botStat, m)
	queryChannel <- BotQuery{"Beholder", m}
}

func (a *Application) pomHandler(m *tbot.Message) {
	logUser(botStat, m)
	// Save time of the request
	updateTime := time.Now()
	// If pom.jpg wasn't updated in an hour do an update
	if pom.Updated.Hour()-updateTime.Hour() != 0 {
		err := updatePomImage(pom.ImageArgs)
		switch {
		// in case there was an error running xplanets send this error as a message
		case err != nil:
			pom.Text = err.Error()
		// otherwise update pom.Text and save the update timestamp
		default:
			pom.Text = getPomText()
			pom.Updated = updateTime
		}
	}
	// Send the image back to Telegram with pom.Text as a caption
	app.Telegram.Client.SendPhotoFile(m.Chat.ID, "pom.jpg", tbot.OptCaption(pom.Text))
}

var ircHandlerFunc = irc.HandlerFunc(func(c *irc.Client, m *irc.Message) {
	switch {
	// Handle WELCOME event
	case m.Command == "001":
		c.Writef("MODE %v -R", app.IRC.Nick)
		// Identify to the NickServ
		c.WriteMessage(&irc.Message{
			Command: "PRIVMSG",
			Params:  []string{"NickServ", app.IRC.Nick, app.IRC.Pass},
		})
	// Handle PING command
	case m.Command == "PING":
		c.Write("PONG")
	// Write private messages from trusted senders to the responseChannel to be picked up by queryWorker
	case m.Command == "PRIVMSG" && isAllowedBot(m.Name, &app):
		responseChannel <- m.Trailing()
	default:
		log.Println(m.Command, m.Params)
	}
},
)
