package lottery_test

import (
	"log"

	"github.com/northbright/lottery-go/lottery"
)

func Example() {
	var (
		err        error
		csvFile    = "participants.example.csv"
		configFile = "config.example.json"
	)

	// Create a lottery.
	l := lottery.New()

	// Load participants from CSV file.
	// CSV file:
	// 1. No headers
	// 2. Contains 2 columns: ID(string) and Name(string)
	if err = l.LoadParticipants(csvFile); err != nil {
		log.Printf("load participants from CSV error: %v\n", err)
		return
	}
	log.Printf("load participants successfully\n")

	participants := l.GetParticipants()
	for _, p := range participants {
		log.Printf("ID: %v, Name: %v\n", p.ID, p.Name)
	}

	// Load config from JSON file.
	if err = l.LoadConfig(configFile); err != nil {
		log.Printf("load config from JSON error: %v\n", err)
		return
	}
	log.Printf("load config successfully\n")

	config := l.GetConfig()
	log.Printf("prizes:\n")
	for _, prize := range config.Prizes {
		log.Printf("name: %v, count: %v, content: %v\n", prize.Name, prize.Num, prize.Content)
	}

	log.Printf("blacklists:\n")
	for _, blacklist := range config.Blacklists {
		log.Printf("max prize index: %v, IDs: %v\n", blacklist.MaxPrizeIndex, blacklist.IDs)
	}

	// Draw a prize.
	winners, err := l.Draw(0)
	log.Printf("Winners: %v\n", winners)

	winners = l.GetWinners(0)
	log.Printf("Winners: %v\n", winners)

	// Output:
}
