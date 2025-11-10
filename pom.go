package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

// pomRequest stores last pom update time and pom description text
type pomRequest struct {
	UpdatedAt time.Time
	Text      string
}

// PoM variable is used to store phase of moon data
var PoM pomRequest

var phaseNames = map[int]string{
	0: "New Moon",
	1: "Waxing Crescent",
	2: "First Quarter",
	3: "Waxing Gibbous",
	4: "Full Moon",
	5: "Waning Gibbous",
	6: "Last Quarter",
	7: "Waning Crescent",
}

// getPhase is a rewrite of the same function from https://alt.org/nethack/moon/pom.pl
// which in turn is a rewrite of NetHack's phase_of_the_moon function
func getPhase(diy, year int) int {
	goldn := (year % 19) + 1
	epact := (11*goldn + 18) % 30
	if (epact == 25 && goldn > 11) || epact == 24 {
		epact++
	}
	return (((((diy + epact) * 6) + 11) % 177) / 22) & 7
}

// isLeapYear returns 1 if it is a leap year and 0 if it is not
func isLeapYear(year int) int {
	leapFlag := 0
	if year%4 == 0 {
		if year%100 == 0 {
			if year%400 == 0 {
				leapFlag = 1
			} else {
				leapFlag = 1
			}
		} else {
			leapFlag = 1
		}
	} else {
		leapFlag = 1
	}
	return leapFlag
}

// updateText is used to construct string describing current moon phase
// example: The Moon is Waxing Gibbous (60% of Full) Full moon in NetHack in 5 days.
func (p *pomRequest) updateText() {
	var (
		inPhase, nextInPhase bool
		days                 int
	)

	localtime := time.Now()
	hour := localtime.Hour()
	year := localtime.Year()
	diy := localtime.YearDay()

	// Generate moon phase description string
	p.Text = fmt.Sprintf("The Moon is %s.\n", phaseNames[getPhase(diy, year)])

	curPhase := getPhase(diy, year)

	if curPhase == 0 || curPhase == 4 {
		inPhase = true
	} else {
		inPhase = false
	}

	leapYear := isLeapYear(year)

	nextDiy, nextYear, nextPhase, nextInPhase := diy, year, curPhase, inPhase

	// adaptation of do…while cycle from the original script
	for {
		nextDiy++
		days++
		// if nextDiy is 31 Dec of a leap year
		if nextDiy-leapYear == 365 {
			nextDiy = 0
			nextYear++
		}
		nextPhase = getPhase(nextDiy, nextYear)
		if nextPhase == 0 || nextPhase == 4 {
			nextInPhase = true
		} else {
			nextInPhase = false
		}
		if inPhase != nextInPhase {
			break
		}
	}

	// completing the message string with NetHack related info
	switch {
	case curPhase == 0:
		p.Text += "New moon in NetHack "
		if days == 1 {
			p.Text += "until midnight, "
		} else {
			p.Text += fmt.Sprintf("for the next %d days.", days)
		}
	case curPhase == 4:
		p.Text += "Full moon in NetHack "
		if days == 1 {
			p.Text += "until midnight, "
		} else {
			p.Text += fmt.Sprintf("for the next %d days.", days)
		}
	case curPhase < 4:
		p.Text += "Full moon in NetHack "
		if days == 1 {
			p.Text += "at midnight, "
		} else {
			p.Text += fmt.Sprintf("in %d days.", days)
		}
	default:
		p.Text += "New moon in NetHack "
		if days == 1 {
			p.Text += "at midnight, "
		} else {
			p.Text += fmt.Sprintf("in %d days.", days)
		}
	}
	// add hour(s) to the message
	if days == 1 {
		p.Text += fmt.Sprintf("%d hour", 24-hour)
		if 24-hour != 1 {
			p.Text += "s"
		}
		p.Text += " from now."
	}
}

// updateImage runs xplanet command with provided arguments
// and returns an error if there any
func (p *pomRequest) updateImage() error {
	args := []string{"-origin", "earth", "-body", "moon",
		"-num_times", "1", "-output", "pom.jpg", "-geometry", "300x300"}
	c := exec.Command("xplanet", args...)
	return c.Run()
}

func (p *pomRequest) init() {
	// Initialize Phase of Moon structure and update pom.jpg
	p.UpdatedAt = time.Now()
	p.updateText()
	err := p.updateImage()
	if err != nil {
		log.Fatal(err)
	}
}
