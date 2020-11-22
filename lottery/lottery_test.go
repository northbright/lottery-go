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
		//configFile = "./config.example.json"
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
