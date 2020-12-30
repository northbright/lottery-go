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

	if err := l.LoadParticipantsCSVFile(participantsCSV); err != nil {
		log.Printf("LoadParticipantsCSVFile() error: %v", err)
		return
	}
	log.Printf("LoadParticipantsCSVFile() successfully")

	log.Printf("participants:")
	participants := l.Participants()
	for _, p := range participants {
		log.Printf("ID: %v, Name: %v", p.ID, p.Name)
	}

	if err := l.LoadPrizesCSVFile(prizesCSV); err != nil {
		log.Printf("LoadPrizesCSVFile() error: %v", err)
		return
	}
	log.Printf("LoadPrizesCSVFile() successfully")

	log.Printf("prizes:")
	prizes := l.Prizes()
	for _, prize := range prizes {
		log.Printf("no: %v, name: %v, count: %v, desc: %v", prize.No, prize.Name, prize.Amount, prize.Desc)
	}

	if err := l.LoadBlacklistsJSONFile(blacklistsJSON); err != nil {
		log.Printf("LoadBlacklistsJSONFile() error: %v", err)
		return
	}
	log.Printf("LoadBlacklistsJSONFile() successfully")

	log.Printf("blacklists:")
	blacklists := l.Blacklists()
	for _, blacklist := range blacklists {
		log.Printf("min prize no: %v, IDs: %v", blacklist.MinPrizeNo, blacklist.IDs)
	}

	// Draw prize no.5.
	log.Printf("draw prize no.5: %v", l.Prize(5))
	winners, err := l.Draw(5)
	if err != nil {
		log.Printf("draw() error: %v", err)
		return
	}

	log.Printf("winners of prize no.5: %v", winners)

	// Revoke old winners and redraw.
	revokedWinners := []lottery.Participant{winners[0], winners[1]}
	if err := l.Revoke(5, revokedWinners); err != nil {
		log.Printf("revoke winners of prize no.5 error: %v", err)
		return
	}
	log.Printf("revoke winners of prize no.5: %v successfully", revokedWinners)

	log.Printf("re-draw prize no.5(amount = 2): %v", l.Prize(5))
	newWinners, err := l.Redraw(5, 2)
	if err != nil {
		log.Printf("Redraw() error: %v", err)
		return
	}
	log.Printf("new winners of prize no.5: %v", newWinners)

	// Get complete updated winners.
	winners = l.Winners(5)
	log.Printf("winners of prize no.5: %v", winners)

	// Draw a prize no.4.
	log.Printf("draw prize no.4: %v", l.Prize(4))
	winners, err = l.Draw(4)
	if err != nil {
		log.Printf("draw() error: %v", err)
		return
	}
	log.Printf("winners of prize no.4: %v", winners)

	// Get all winners.
	log.Printf("get all winners:")
	allWinners := l.AllWinners()
	for no, winners := range allWinners {
		log.Printf("prize no %v: %v", no, winners)
	}

	// Save data(include all winners).
	if err := l.SaveToFile(); err != nil {
		log.Printf("Save() error: %v", err)
		return
	}
	log.Printf("save data successfully")

	// Load data
	if err := l.LoadFromFile(); err != nil {
		log.Printf("Load() error: %v", err)
		return
	}
	log.Printf("load data successfully")

	// Clear winners for prize no == 5
	l.ClearWinners(5)
	log.Printf("clear winners of prize 5")

	// Save data
	if err := l.SaveToFile(); err != nil {
		log.Printf("Save() error: %v", err)
		return
	}
	log.Printf("save data successfully")

	// Load data
	if err := l.LoadFromFile(); err != nil {
		log.Printf("Load() error: %v", err)
		return
	}
	log.Printf("load data successfully")

	// Get all winners again.
	allWinners = l.AllWinners()
	for no, winners := range allWinners {
		log.Printf("prize no %v: %v", no, winners)
	}

	// Output:
}
