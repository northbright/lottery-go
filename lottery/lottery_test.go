package lottery_test

import (
	"log"

	"github.com/northbright/lottery-go/lottery"
)

func Example() {
	var (
		participantsCSV = "participants.example.csv"
		prizesCSV       = "prizes.example.csv"
		blacklistsJSON  = "blacklists.example.json"
	)

	// Create a lottery.
	l := lottery.New("New Year Party Lucky Draw")

	if err := l.SetParticipantsFromCSV(participantsCSV); err != nil {
		log.Printf("SetParticipantsFromCSV() error: %v", err)
		return
	}
	log.Printf("SetParticipantsFromCSV() successfully")

	log.Printf("participants:")
	for _, p := range l.Participants {
		log.Printf("ID: %v, Name: %v", p.ID, p.Name)
	}

	if err := l.SetPrizesFromCSV(prizesCSV); err != nil {
		log.Printf("SetPrizesFromCSV() error: %v", err)
		return
	}
	log.Printf("SetPrizesFromCSV() successfully")

	log.Printf("prizes:")
	for prizeNo, prize := range l.Prizes {
		log.Printf("no: %v, name: %v, count: %v, desc: %v", prizeNo, prize.Name, prize.Amount, prize.Desc)
	}

	if err := l.SetBlacklistsFromJSON(blacklistsJSON); err != nil {
		log.Printf("SetBlacklistsFromJSON() error: %v", err)
		return
	}
	log.Printf("SetBlacklistsFromJSON() successfully")

	log.Printf("blacklists:\n")
	for maxPrizeNo, blacklist := range l.Blacklists {
		log.Printf("max prize no: %v, IDs: %v", maxPrizeNo, blacklist.IDs)
	}

	// Draw prize no.5.
	log.Printf("draw prize no.5: %v", l.Prizes[5])
	winners, err := l.Draw(5)
	if err != nil {
		log.Printf("draw() error: %v", err)
		return
	}

	log.Printf("winners of prize no.5: %v", winners)

	// Revoke old winners and redraw.
	revokedWinners := []lottery.Participant{winners[0], winners[1]}
	log.Printf("revoke winners of prize no.5: %v", revokedWinners)

	log.Printf("re-draw prize no.5: %v", l.Prizes[5])
	newWinners, err := l.Redraw(5, revokedWinners)
	if err != nil {
		log.Printf("Redraw() error: %v", err)
		return
	}
	log.Printf("new winners of prize no.5: %v", newWinners)

	// Get complete updated winners.
	winners = l.GetWinners(5)
	log.Printf("winners of prize no.5: %v", winners)

	// Draw a prize no.4.
	log.Printf("draw prize no.4: %v", l.Prizes[4])
	winners, err = l.Draw(4)
	if err != nil {
		log.Printf("draw() error: %v", err)
		return
	}
	log.Printf("winners of prize no.4: %v", winners)

	// Get all winners.
	log.Printf("get all winners:")
	allWinners := l.GetAllWinners()
	for no, winners := range allWinners {
		log.Printf("prize no %v: %v", no, winners)
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

	// Clear winners for prize no == 5
	l.ClearWinners(5)
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

	// Get all winners again.
	allWinners = l.GetAllWinners()
	for no, winners := range allWinners {
		log.Printf("prize no %v: %v", no, winners)
	}

	log.Printf("l: %v", l)
	log.Printf("l.Participants: %v", l.Participants)

	// Output:
}
