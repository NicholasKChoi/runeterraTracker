package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

	cardCodeToName = make(map[string]string)
)

type StaticDeckList struct {
	DeckCode    string
	CardsInDeck map[string]int
}

type Cards []struct {
	CardCode    string
	CardName    string `json:"name"`
}

func fillCardCodeToNameMap() error{
	jsonFile, err := os.Open("../en_us/data/set1-en_us.json")
	if err != nil {
		return errors.Wrap(err, "Unable to set json file.")
	}

	cards := Cards{}
	dec := json.NewDecoder(jsonFile)
	if err = dec.Decode(&cards); err != nil {
		return errors.Wrap(err, "json file could not be interpreted as a set list.")
	}

	for _, item := range cards {
		cardCodeToName[item.CardCode] = item.CardName
	}
	return nil
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

func convertDeckListCodesToNames(list *StaticDeckList) StaticDeckList{
	deckTracker := make(map[string]int)
	for key, value := range list.CardsInDeck {
		cardName := cardCodeToName[key]
		deckTracker[cardName] = value
	}
	list.CardsInDeck = deckTracker
	return *list
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
				currentDeckList = convertDeckListCodesToNames(&newdecklist)
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
	if err := fillCardCodeToNameMap(); err != nil {
		log.Fatal(err)
	}
	args := mainLoopArgs{}
	mainloop(args)
}
