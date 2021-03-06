package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

/**
{
  "DeckCode": "DECKCODE",
  "CardsInDeck": {
    "01DE000": 1,
    "01DE001": 2,
    ...
  },
  "garbage": 10000001
}

*/

var (
	// the port number that legends is available on in the local host -- this is currently set ot the default value
	// todo make this variable update later on
	legendsPort = 21337

	// this is the base pf the url that the legend host can be found on
	// todo udpate the legendHost based on the legendsPort (when it changes)
	// url.Parse -- but I personally find this code hard to read.
	legendHost = fmt.Sprintf("http://localhost:%d", legendsPort)
)

type StaticDeckList struct {
	DeckCode    string
	CardsInDeck map[string]int
}

func getDeckList() (StaticDeckList, error) {
	sdl := StaticDeckList{}

	// fetch the response from the legends of runterra api
	endpoint := fmt.Sprintf("%s/static-decklist", legendHost)
	resp, err := http.Get(endpoint)
	if err != nil {
		return sdl, errors.Wrap(err, "getting static decklist failed")
	}

	// decode the response into the structure that we want to fill out
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&sdl)
	if err != nil {
		return sdl, errors.Wrap(err, "response could not be interpreted as a static deck list")
	}
	return sdl, nil
}

type mainLoopArgs struct {
}

func mainloop(args mainLoopArgs) {
	var (
		signalUpdateDecklist = time.NewTicker(4 * time.Second)
		currentDeckList      = StaticDeckList{}
	)

	for {
		select {
		case <-signalUpdateDecklist.C:
			newdecklist, err := getDeckList()
			if err != nil {
				// %+v prints error with the back trace
				fmt.Println("encountered error fetching new decklist value:\n", err)
			} else {
				currentDeckList = newdecklist
				spew.Dump(currentDeckList)
			}

		}

		// state processing
		if currentDeckList.DeckCode != "" {
			fmt.Println("I'm in game!")
		}
	}
}

func main() {
	// this tells log to print with the short version of the file and line number and the time
	log.SetFlags(log.Lshortfile | log.Ltime)

	// makes sure that we can make any calls to the legend local server -- should fail if runeterra is closed
	//resp, err := http.Get(legendHost)
	//if err != nil {
	//	log.Println(err)
	//}
	//bytz, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("Contents:\n%s", string(bytz))

	// start the mainloop of the deck tracker
	args := mainLoopArgs{}
	mainloop(args)
}
