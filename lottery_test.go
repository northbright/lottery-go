package lottery_test

import (
	"log"

	"github.com/northbright/lottery-go/lottery"
)

func Example() {
	var (
		err        error
		csvFile    = "./participants.example.csv"
		configFile = "./config.example.json"
	)

	l := lottery.Lottery{}
	if err = l.LoadParticipants(csvFile); err != nil {
		log.Printf("load participants from CSV error: %v\n", err)
		return
	}

	// Output:
}
