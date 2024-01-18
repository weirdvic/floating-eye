package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/mroth/weightedrand/v2"
	"github.com/yanzay/tbot/v2"
	"gopkg.in/irc.v3"
)

type application struct {
	Config struct {
		Players []string `json:"watch_players"`
	} `json:"config"`
	Telegram struct {
		Client      *tbot.Client
		Token       string `json:"token"`
		Admins      []int  `json:"admins"`
		ForwardChat int    `json:"forward_chat"`
	} `json:"telegram"`
	IRC struct {
		Client *irc.Client
		Server string   `json:"server"`
		Port   int      `json:"port"`
		Nick   string   `json:"nick"`
		Pass   string   `json:"pass"`
		Name   string   `json:"name"`
		Bots   []string `json:"bots"`
	} `json:"irc"`
	Filters map[string]*regexp.Regexp
	Potions map[string]string
}

type botQuery struct {
	BotNick string
	Query   *tbot.Message
}

// askBot is a simple wrapper to send message to the IRC bot
func askBot(nick, text string) {
	app.IRC.Client.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params:  []string{nick, text}})
}

// getMonsterName parses string to extract monster's name
func getMonsterName(r *regexp.Regexp, s string) (name string, e error) {
	if !r.MatchString(s) {
		return "", errors.New("provided string does not contain a monster name")
	}
	match := r.FindStringSubmatch(s)
	if match[2] == "" {
		name = match[1]
		return name, nil
	}
	// else
	name = match[2]
	return name, nil
}

// checkBotName checks if bot name is in allowed list
func (a *application) checkBotName(item string) bool {
	for _, v := range a.IRC.Bots {
		if item == v {
			return true
		}
	}
	return false
}

// parseChatMessage parses IRC message and if it's mentions one of the players,
// forward that message to Telegram
func (a *application) parseChatMessage(m string) {
	if !app.Filters["mentions"].MatchString(m) {
		return
	} else {
		app.Telegram.Client.SendMessage(
			strconv.Itoa(app.Telegram.ForwardChat),
			app.Filters["TGannounce"].ReplaceAllString(m, `[$1]`),
		)
	}
}

// queryWorker reads from inboxChannel, passes the query text to IRC,
// awaits for response from bot and sends the response text back to Telegram
func queryWorker(c <-chan botQuery) {
	for q := range c {
		askBot(q.BotNick, q.Query.Text)
		// Read response from the channel
		botResponse := <-responseChannel
		// Filter IRC color codes and replace parentheses to brackets
		botResponse = app.Filters["IRCcolors"].ReplaceAllString(botResponse, `[ $1$2 ]`)
		// Split response to lines by '|' symbol
		botResponse = strings.ReplaceAll(botResponse, "|", "\n")
		// In case we're working on monster query
		if q.BotNick == "Pinoclone" {
			// Parsing monster's name
			monsterName, err := getMonsterName(app.Filters["monsterName"], botResponse)
			if err == nil {
				fileName := fmt.Sprintf("monsters/%s.png", strings.ToUpper(monsterName))
				// If monster's image is not available, set placeholder image
				if _, err = os.Stat(fileName); err != nil {
					log.Println("Image not found: ", fileName)
					fileName = "monsters/WARNING 0.png"
				}
				// Send image with caption
				app.Telegram.Client.SendPhotoFile(
					q.Query.Chat.ID,
					fileName,
					tbot.OptCaption(
						strings.Join([]string{botResponse,
							"https://nethackwiki.com/wiki/" + strings.Title(
								strings.ToLower(
									strings.ReplaceAll(monsterName, " ", "_")))},
							"\n"),
					),
					tbot.OptReplyToMessageID(q.Query.MessageID),
				)
			} else {
				log.Println(err)
				app.Telegram.Client.SendMessage(q.Query.Chat.ID, botResponse)
			}
		} else {
			// Send just text for other queries
			app.Telegram.Client.SendMessage(q.Query.Chat.ID, botResponse)
		}
	}
}

// readConfig decodes embedded config.json file to struct
func (a *application) init() {
	// Check if config.json file exists
	_, err := os.Stat("config.json")
	if os.IsNotExist(err) {
		log.Fatal("config.json file does not exist!")
	}
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal("Could not read config.json!")
	}
	// Decode config.json file to app struct
	err = json.Unmarshal(configFile, a)
	if err != nil {
		log.Fatal(err)
	}
	// Init filters
	if app.Filters == nil {
		app.Filters = make(map[string]*regexp.Regexp)
	}

	// Brewing potions
	if app.Potions == nil {
		app.Potions = brewPotions()
	}

	log.Println("Config successfully loaded.")
	log.Print("Available bots are: ")
	log.Println(a.IRC.Bots)
	// Check if dependency commands are available in the system
	var commands = []string{"pom", "xplanet"}
	for _, v := range commands {
		path, err := exec.LookPath(v)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Command %s is located at %s\n", v, path)
	}
	// Check if "monsters" directory exists
	_, err = os.Stat("monsters")
	if os.IsNotExist(err) {
		log.Fatal("Monster images directory does not exist!")
	}

	// This regexp is used to filter IRC color codes from Pinoclone's response
	app.Filters["IRCcolors"] = regexp.MustCompile(
		`\(.*\d{1,2},\d{1,2}(\S).*\)|\[\s+\d{1,2}(\S+)\s+\]`)
	app.Filters["TGannounce"] = regexp.MustCompile(
		`\[\D*\d{1,2}([a-zA-Z|-]+)\D*\]`)
	app.Filters["monsterName"] = regexp.MustCompile(
		`^([\w+\s+-]+)\s\[|~\d+~\s([\w+\s+-]+)\s\[`)

	// Construct regexp to find messages that mention players
	app.Filters["mentions"] = regexp.MustCompile(
		fmt.Sprint(
			`^.*\]:?\s(`,
			strings.Join(app.Config.Players, `|`),
			`)\s\(.*$`))

	log.Println("All checks passed…")
}

// Send message to admin on shutdown
func (a *application) shutdown(reason string) {
	a.Telegram.Client.SendMessage(
		strconv.Itoa(a.Telegram.Admins[0]),
		fmt.Sprintf("Shutting down on: %s", reason))
	a.IRC.Client.Write("QUIT")
}

func makeOrcName() string {
	var s string
	v := [...]string{"a", "ai", "og", "u"}
	snd := [...]string{"gor", "gris", "un", "bane", "ruk", "oth", "ul", "z", "thos", "akh", "hai"}

	iend := rand.Intn(2) + 3
	vstart := rand.Intn(2)

	for i := 0; i < iend; i++ {
		vstart += -1 /* 0 -> 1, 1 -> 0 */
		if i > 0 && rand.Intn(30) == 0 {
			s += "-"
		}
		if vstart == 1 {
			s += v[rand.Intn(len(v))]
		} else {
			s += snd[rand.Intn(len(snd))]
		}
	}
	return s
}

// brewPotions returns a map of potion effects to randomized appearances
func brewPotions() map[string]string {
	effects := []string{
		"booze", "fruit juice", "see invisible", "sickness", "confusion",
		"extra healing", "hallucination", "healing", "restore ability", "sleeping",
		"blindness", "gain energy", "invisibility", "monster detection", "object detection",
		"enlightenment", "full healing", "levitation", "polymorph", "speed",
		"acid", "oil", "gain ability", "gain level", "paralysis",
	}
	appearances := []string{
		"ruby", "pink", "orange", "yellow", "emerald",
		"dark green", "cyan", "sky blue", "brilliant blue", "magenta",
		"purple-red", "puce", "milky", "swirly", "bubbly",
		"smoky", "cloudy", "effervescent", "black", "golden",
		"brown", "fizzy", "dark", "white", "murky",
	}
	rand.Shuffle(
		len(appearances),
		func(i, j int) { appearances[i], appearances[j] = appearances[j], appearances[i] },
	)

	potions := make(map[string]string)
	potions["water"] = "clear"
	potions["holy water"] = "clear"
	potions["unholy water"] = "clear"

	for i, effect := range effects {
		potions[effect] = appearances[i]
	}
	return potions
}

// pickPotion returns a random potion effect
func pickPotion() string {
	potionSeller, _ := weightedrand.NewChooser(
		weightedrand.NewChoice("water", 690),
		weightedrand.NewChoice("holy water", 115),
		weightedrand.NewChoice("unholy water", 115),
		weightedrand.NewChoice("booze", 420),
		weightedrand.NewChoice("fruit juice", 420),
		weightedrand.NewChoice("see invisible", 420),
		weightedrand.NewChoice("sickness", 420),
		weightedrand.NewChoice("confusion", 420),
		weightedrand.NewChoice("extra healing", 470),
		weightedrand.NewChoice("hallucination", 400),
		weightedrand.NewChoice("healing", 570),
		weightedrand.NewChoice("restore ability", 400),
		weightedrand.NewChoice("sleeping", 420),
		weightedrand.NewChoice("blindness", 400),
		weightedrand.NewChoice("gain energy", 420),
		weightedrand.NewChoice("invisibility", 400),
		weightedrand.NewChoice("monster detection", 400),
		weightedrand.NewChoice("object detection", 420),
		weightedrand.NewChoice("enlightenment", 200),
		weightedrand.NewChoice("full healing", 100),
		weightedrand.NewChoice("levitation", 420),
		weightedrand.NewChoice("polymorph", 100),
		weightedrand.NewChoice("speed", 420),
		weightedrand.NewChoice("acid", 100),
		weightedrand.NewChoice("oil", 300),
		weightedrand.NewChoice("gain ability", 420),
		weightedrand.NewChoice("gain level", 200),
		weightedrand.NewChoice("paralysis", 420),
		weightedrand.NewChoice("nothing", 1),
	)
	return potionSeller.Pick()
}
