package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

// scrapePOM gets current Lunar phase and corresponding image from https://alt.org/nethack site
func scrapePOM() (pomFileName, pomText string) {
	var (
		fileHandler *os.File
		pomTextURL  = "https://alt.org/nethack/"
		pomImageURL = "https://alt.org/nethack/moon/pom.jpg"
	)
	// Getting Lunar phase description text
	res, err := http.Get(pomTextURL)
	checkError(err)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		// Phase of moon is stored in second <p></p> section so we use 1 as index and save section's tex to the variable
		if i == 1 {
			pomText = s.Text()
		}
	})

	// Getting Lunar phase image
	res, err = http.Get(pomImageURL)
	checkError(err)
	defer res.Body.Close()
	// We are saving response body to pomImage to be able to read it several times
	// res.Body not used after this line
	pomImage, err := ioutil.ReadAll(res.Body)

	// First of all we're going to calculate md5 hash of a file to use it as a filename
	hash := md5.New()
	_, err = io.Copy(hash, ioutil.NopCloser(bytes.NewReader(pomImage)))
	checkError(err)
	md5sum := hex.EncodeToString(hash.Sum(nil))
	// filename is pom/md5sum.jpg
	pomFileName = "pom/" + md5sum + ".jpg"

	// Let's check if file with that name already exists
	_, err = os.Stat(pomFileName)
	// Create new file if pomFileName not found
	if err != nil {
		log.Println("File not found: ", pomFileName)
		fileHandler, err = os.Create(pomFileName)
		checkError(err)
		defer fileHandler.Close()
		// Copying response body bytes to file
		_, err = io.Copy(fileHandler, ioutil.NopCloser(bytes.NewReader(pomImage)))
		checkError(err)
		log.Println("Image was saved in: ", pomFileName)
		// Return the filename of image with current Lunar phase and description string
		return pomFileName, pomText
	}
	// In case file already exist return filename and description string
	return pomFileName, pomText
}

// checkError is a simple wrapper for "if err != nil" construction
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
