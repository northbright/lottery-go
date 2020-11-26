package lottery_test

import (
	"fmt"
	"log"

	"github.com/northbright/lottery-go/lottery"
)

func ExampleLottery_LoadParticipants() {
	var (
		err     error
		csvFile = "participants.example.csv"
	)

	l := lottery.Lottery{}
	if err = l.LoadParticipants(csvFile); err != nil {
		log.Printf("load participants from CSV error: %v\n", err)
		return
	}

	fmt.Printf("load participants successfully\n")

	participants := l.GetParticipants()
	for _, p := range participants {
		fmt.Printf("ID: %v, Name: %v\n", p.ID, p.Name)
	}

	// Output:
	//load participants successfully
	//ID: 5, Name: Fal
	//ID: 7, Name: Nango
	//ID: 8, Name: Jacky
	//ID: 9, Name: Sonny
	//ID: 10, Name: Luke
	//ID: 11, Name: Mic
	//ID: 12, Name: Ric
	//ID: 13, Name: Capt
	//ID: 14, Name: Andy
	//ID: 17, Name: Alex
	//ID: 33, Name: Xiao
}

func ExampleLottery_LoadConfig() {
	var (
		err        error
		configFile = "config.example.json"
	)

	l := lottery.Lottery{}
	if err = l.LoadConfig(configFile); err != nil {
		log.Printf("load config from JSON error: %v\n", err)
		return
	}

	fmt.Printf("load config successfully\n")

	config := l.GetConfig()
	fmt.Printf("prizes:\n")
	for _, prize := range config.Prizes {
		fmt.Printf("name: %v, count: %v, content: %v\n", prize.Name, prize.Num, prize.Content)
	}

	fmt.Printf("blacklists:\n")
	for _, blacklist := range config.Blacklists {
		fmt.Printf("max prize index: %v, IDs: %v\n", blacklist.MaxPrizeIndex, blacklist.IDs)
	}

	// Output:
	//load config successfully
	//prizes:
	//name: grade 5 prize, count: 10, content: USB Hard drive
	//name: grade 4 prize, count: 8, content: Bluetooth Speaker
	//name: grade 3 prize, count: 5, content: Vacuum Cleaner
	//name: grade 2 prize, count: 2, content: Macbook Pro
	//name: grade 1 prize, count: 1, content: iPhone
	//blacklists:
	//max prize index: 2, IDs: [33]
}
