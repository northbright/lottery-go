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
		log.Printf("name: %v, count: %v, desc: %v", prize.Name, prize.Num, prize.Desc)
	}

	log.Printf("blacklists:\n")
	for _, blacklist := range config.Blacklists {
		log.Printf("max prize index: %v, IDs: %v", blacklist.MaxPrizeIndex, blacklist.IDs)
	}

	// Draw a prize(index=0, grade=5).
	nPrizeIndex := 0
	log.Printf("draw prize %v: %v", nPrizeIndex, config.Prizes[nPrizeIndex])
	winners, _ := l.Draw(nPrizeIndex)
	log.Printf("winners of prize %v: %v", nPrizeIndex, winners)

	// Revoke old winners and redraw.
	revokedWinners := []lottery.Participant{winners[0], winners[1]}
	log.Printf("revoke winners of prize %v: %v", nPrizeIndex, revokedWinners)

	newWinners, _ := l.Redraw(nPrizeIndex, revokedWinners)
	log.Printf("re-draw prize %v: %v", nPrizeIndex, config.Prizes[nPrizeIndex])
	log.Printf("new winners of prize %v: %v", nPrizeIndex, newWinners)

	// Get complete updated winners.
	winners = l.GetWinners(nPrizeIndex)
	log.Printf("winners of prize %v: %v", nPrizeIndex, winners)

	// Draw a prize(index=1, grade=4).
	nPrizeIndex = 1
	log.Printf("draw prize %v: %v", nPrizeIndex, config.Prizes[nPrizeIndex])
	winners, _ = l.Draw(nPrizeIndex)
	log.Printf("winners of prize %v: %v", nPrizeIndex, winners)

	// Get all winners.
	allWinners := l.GetAllWinners()
	for i, winners := range allWinners {
		log.Printf("prize index %v: %v", i, winners)
	}

	// Save data(include all winners).
	if err := l.Save(); err != nil {
		log.Printf("Save() error: %v", err)
		return
	}
	log.Printf("save data successfully")

	// Load data
	if err := l.Load(); err != nil {
		log.Printf("Load() error: %v", err)
		return
	}
	log.Printf("load data successfully")

	// Clear winners for prize index == 0
	l.ClearWinners(0)
	log.Printf("clear winners of prize 0")

	// Save data
	if err := l.Save(); err != nil {
		log.Printf("Save() error: %v", err)
		return
	}
	log.Printf("save data successfully")

	// Load data
	if err := l.Load(); err != nil {
		log.Printf("Load() error: %v", err)
		return
	}
	log.Printf("load data successfully")

	// Get winners of prize index == 0 again.
	winners = l.GetWinners(0)
	log.Printf("winners of prize 0: %v", winners)

	// Output:
}
