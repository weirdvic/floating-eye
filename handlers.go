package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/yanzay/tbot/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	a.Telegram.Client.SendMessage(m.Chat.ID, cases.Title(language.English).String(makeOrcName()))
}

// Handler for !quaff command
func (a *application) quaffHandler(m *tbot.Message) {
	potion := pickPotion()
	if app.Potions[potion] == "milky" && rand.Intn(100) < 7 {
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. As you open the bottle, an enormous ghost emerges! You are frightened to death, and unable to move.", app.Potions[potion]),
			tbot.OptReplyToMessageID(m.MessageID))
		return
	} else if app.Potions[potion] == "smoky" && rand.Intn(100) < 7 {
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. As you open the bottle, an enormous djinni emerges!", app.Potions[potion]),
			tbot.OptReplyToMessageID(m.MessageID))
		return
	}
	switch potion {
	case "nothing":
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			"You mime drinking something",
			tbot.OptReplyToMessageID(m.MessageID))
	case "water":
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. This tastes like water", app.Potions[potion]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "holy water":
		messages := []string{
			"You feel full of awe.",
			"This burns like acid!",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "unholy water":
		messages := []string{
			"This burns like acid!",
			"You feel full of dread.",
			"You feel quite proud of yourself.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "booze":
		messages := []string{
			"Ooph! This tastes like liquid fire!",
			"Ooph! This tastes like watered down liquid fire!",
			"Ooph! This tastes like dandelion wine!",
			"Ooph! This tastes like watered down dandelion wine!",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "fruit juice":
		messages := []string{
			"This tastes like slime mold juice.",
			"Yecch! This tastes rotten.",
			"This tastes like reconstituted slime mold juice.",
			"This tastes like 10% real slime mold juice all-natural beverage.",
			"This tastes like 10% real reconstituted <slime mold> juice all-natural beverage.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "see invisible":
		messages := []string{
			"You can see through yourself, but you are visible!",
			"Gee! All of a sudden, you can see right through yourself.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "sickness":
		messages := []string{
			"(But in fact it was mildly stale slime mold juice.)",
			"(But in fact it was biologically contaminated slime mold juice.)",
			"Fortunately, you have been immunized.",
			"You are shocked back to your senses!",
			"You feel weaker.",
			"Your muscles won't obey you.",
			"You feel very sick.",
			"Your brain is on fire.",
			"Your judgement is impaired.",
			"You break out in hives.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. Yecch! This stuff tastes like poison. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "confusion":
		messages := []string{
			"Huh, What? Where am I?",
			"What a trippy feeling!",
			"You feel somewhat dizzy",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "extra healing":
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. You feel much better.", app.Potions[potion]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "hallucination":
		messages := []string{
			"Oh wow! Everything looks so cosmic!",
			"You have a normal feeling for a moment, then it passes.",
			"Your vision seems to flatten for a moment but is normal now.",
			"Your eyes itch.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "healing":
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. You feel better.", app.Potions[potion]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "full healing":
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. You feel completely healed.", app.Potions[potion]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "restore ability":
		messages := []string{
			"Wow! This makes you feel great!",
			"Wow! This makes you feel better!",
			"Ulch! This makes you feel mediocre!",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "sleeping":
		messages := []string{
			"You fall asleep.",
			"You yawn.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "blindness":
		messages := []string{
			"A cloud of darkness falls upon you.",
			"Oh, bummer! Everything is dark! Help!",
			"Your eyes itch.",
			"You have a strange feeling for a moment, then it passes.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "gain energy":
		messages := []string{
			"Magical energies course through your body.",
			"You feel lackluster.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "invisibility":
		messages := []string{
			"You have a peculiar feeling for a moment, then it passes.",
			"You have a normal feeling for a moment, then it passes.",
			"Gee! All of a sudden, you can't see yourself.",
			"Far out, man! You can't see yourself.",
			"Gee! All of a sudden, you can see right through yourself.",
			"Far out, man! You can see right through yourself.",
			"For some reason, you feel your presence is known.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "monster detection":
		messages := []string{
			"You can sense the presence of monsters.",
			"Monsters sense the presence of you.",
			"You feel threatened.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "object detection":
		messages := []string{
			"You sense the presence of objects.",
			"You sense the presence of something.",
			"You sense the absence of objects.",
			"You sense something nearby.",
			"You sense objects nearby.",
			"You feel a lack of something.",
			"You have a normal feeling for a moment, then it passes.",
			"You have a strange feeling for a moment, then it passes.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "enlightenment":
		messages := []string{
			"You have an uneasy feeling…",
			"You feel self-knowledgeable…",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "levitation":
		messages := []string{
			"You have a peculiar feeling for a moment, then it passes.",
			"You have a normal feeling for a moment, then it passes.",
			"You hit your head on the ceiling.",
			"You start to float in the air!",
			"Up, up, and awaaaay! You're walking on air!",
			"You float up, out of the pit!",
			"Your body pulls upward, but your legs are still stuck.",
			"You float up, only your leg is still stuck.",
			"You float away from the maw.",
			"You gain control over your movements.",
			"You are no longer able to control your flight.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "polymorph":
		messages := []string{
			"You feel a little strange. You turn into a zruty!",
			"You feel a little normal. You turn into a wumpus!",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "speed":
		messages := []string{
			"You are suddenly moving much faster.",
			"Your legs get new energy.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "acid":
		messages := []string{
			"This burns a lot!",
			"This burns like acid!",
			"This tastes sour.",
			"This burns!",
			"This burns a little!",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "oil":
		messages := []string{
			"Ahh, a refreshing drink.",
			"That was smooth!",
			"This tastes like castor oil.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "gain ability":
		messages := []string{
			"You feel strong!",
			"You feel smart!",
			"You feel wise!",
			"You feel agile!",
			"You feel tough!",
			"You feel charismatic!",
			"You're already as strong as you can get.",
			"You're already as smart as you can get.",
			"You're already as wise as you can get.",
			"You're already as agile as you can get.",
			"You're already as tough as you can get.",
			"You're already as charismatic as you can get.",
			"Ulch! That potion tasted foul!",
			"You have a peculiar feeling for a moment, then it passes.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "gain level":
		messages := []string{
			"You feel more experienced.",
			"You rise up, through the ceiling!",
			"You have an uneasy feeling.",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	case "paralysis":
		messages := []string{
			"Your feet are frozen to the floor!",
			"You stiffen momentarily!",
			"You are motionlessly suspended.",
			"You are frozen in place!",
		}
		a.Telegram.Client.SendMessage(
			m.Chat.ID,
			fmt.Sprintf("You drink %s potion. %s", app.Potions[potion], messages[rand.Intn(len(messages))]),
			tbot.OptReplyToMessageID(m.MessageID))
	}
}
