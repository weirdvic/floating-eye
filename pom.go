package main

import (
	"fmt"
	"os/exec"
	"time"
)

// pomRequest stores last pom update time, pom description text and xplanet arguments
type pomRequest struct {
	Updated   time.Time
	Text      string
	ImageArgs []string
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

// getPomText is used to construct string describing current moon phase
// example: The Moon is Waxing Gibbous (60% of Full) Full moon in NetHack in 5 days.
func getPomText() (pomText string) {
	var (
		inPhase, days int
	)
	// first part of the message is the result of 'pom' command from bsdgames package
	pomOut, err := exec.Command("pom").Output()
	if err != nil {
		pomText = err.Error()
		return pomText
	}
	// appending first part of the message
	pomText = string(pomOut)

	localtime := time.Now()
	hour := localtime.Hour()
	year := localtime.Year()
	diy := localtime.YearDay()

	curPhase := getPhase(diy, year)

	if curPhase == 0 || curPhase == 4 {
		inPhase = 1
	}

	leapYear := isLeapYear(year)

	nextDiy, nextYear, nextPhase, nextInPhase := diy, year, curPhase, inPhase

	// adaptation of do…while cycle from the original script
	for {
		nextDiy++
		days++
		if nextDiy-leapYear == 365 {
			nextDiy = 0
			nextYear++
		}
		nextPhase = getPhase(nextDiy, nextYear)
		if nextPhase == 0 || nextPhase == 4 {
			nextInPhase = 1
		} else {
			nextInPhase = 0
		}
		if inPhase != nextInPhase {
			break
		}
	}

	// completing the message string with NetHack related info
	switch {
	case curPhase == 0:
		pomText += "New moon in NetHack "
		if days == 1 {
			pomText += "until midnight, "
		} else {
			pomText += fmt.Sprintf("for the next %d days.", days)
		}
	case curPhase == 4:
		pomText += "Full moon in NetHack "
		if days == 1 {
			pomText += "until midnight, "
		} else {
			pomText += fmt.Sprintf("for the next %d days.", days)
		}
	case curPhase < 4:
		pomText += "Full moon in NetHack "
		if days == 1 {
			pomText += "at midnight, "
		} else {
			pomText += fmt.Sprintf("in %d days.", days)
		}
	default:
		pomText += "New moon in NetHack "
		if days == 1 {
			pomText += "at midnight, "
		} else {
			pomText += fmt.Sprintf("in %d days.", days)
		}
	}
	// add hour(s) to the message
	if days == 1 {
		pomText += fmt.Sprintf("%d hour", 24-hour)
		if 24-hour != 1 {
			pomText += "s"
		}
		pomText += " from now."
	}
	return pomText
}

// updatePomImage runs xplanet command with provided arguments and returns an error if there any
func updatePomImage(args []string) error {
	c := exec.Command("xplanet", args...)
	return c.Run()
}
