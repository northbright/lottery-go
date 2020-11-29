package lottery_test

import (
	"log"

	"github.com/northbright/lottery-go/lottery"
)

func Example() {
	var (
		csvFile    = "participants.example.csv"
		configFile = "config.example.json"
	)

	// Create a lottery.
	l := lottery.New(csvFile, configFile)
	if l == nil {
		log.Printf("invalid participants csv file or config file")
		return
	}

	participants := l.GetParticipants()
	log.Printf("participants:")
	for _, p := range participants {
		log.Printf("ID: %v, Name: %v", p.ID, p.Name)
	}

	config := l.GetConfig()
	log.Printf("name: %s", config.Name)
	log.Printf("prizes:")
	for _, prize := range config.Prizes {
		log.Printf("name: %v, count: %v, content: %v", prize.Name, prize.Num, prize.Content)
	}

	log.Printf("blacklists:\n")
	for _, blacklist := range config.Blacklists {
		log.Printf("max prize index: %v, IDs: %v", blacklist.MaxPrizeIndex, blacklist.IDs)
	}

	// Draw a prize.
	nPrizeIndex := 0
	log.Printf("draw prize: %v(index = %v)", config.Prizes[nPrizeIndex], nPrizeIndex)
	winners, _ := l.Draw(nPrizeIndex)
	log.Printf("winners: %v", winners)

	// Revoke old winners and redraw.
	revokedWinners := []lottery.Participant{winners[0], winners[1]}
	log.Printf("revoke winners: %v", revokedWinners)

	newWinners, _ := l.Redraw(nPrizeIndex, revokedWinners)
	log.Printf("re-draw prize: %v(index = %v)", config.Prizes[nPrizeIndex], nPrizeIndex)
	log.Printf("new winners: %v", newWinners)

	// Get complete updated winners.
	winners = l.GetWinners(nPrizeIndex)
	log.Printf("winners: %v", winners)

	// Output:
}
