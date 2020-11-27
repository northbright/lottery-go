package lottery

import (
	//"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	//"math/rand"
	//"sync"
	//"time"

	"github.com/northbright/csvhelper"
)

type Participant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Prize struct {
	Name    string `json:"name"`
	Num     int    `json:"num"`
	Content string `json:"content"`
}

type Blacklist struct {
	MaxPrizeIndex int      `json:"max_prize_index"`
	IDs           []string `json:"ids"`
}

type Config struct {
	Prizes     []Prize     `json:"prizes"`
	Blacklists []Blacklist `json:"blacklists"`
}

type Lottery struct {
	config       Config
	participants map[string]Participant
	winners      map[int][]string
}

var (
	ErrParticipantsCSV = fmt.Errorf("incorrect participants CSV")
)

func (l *Lottery) LoadParticipants(file string) error {
	rows, err := csvhelper.ReadFile(file)
	if err != nil {
		return err
	}

	l.participants = make(map[string]Participant)
	for _, row := range rows {
		if len(row) != 2 {
			return ErrParticipantsCSV
		}
		l.participants[row[0]] = Participant{row[0], row[1]}
	}
	return nil
}

func participantsMapToSlice(m map[string]Participant) []Participant {
	var participants []Participant

	for _, p := range m {
		participants = append(participants, p)
	}

	sort.Slice(participants, func(i, j int) bool {
		// Try to convert ID from string to uint
		nID1, err1 := strconv.ParseUint(participants[i].ID, 10, 64)
		nID2, err2 := strconv.ParseUint(participants[j].ID, 10, 64)

		// Compare strings
		if err1 != nil || err2 != nil {
			return participants[i].ID < participants[j].ID
		}

		// Compare uints
		return nID1 < nID2
	})

	return participants
}

func (l *Lottery) GetParticipants() []Participant {
	return participantsMapToSlice(l.participants)
}

func (l *Lottery) LoadConfig(file string) error {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, &l.config)
}

func (l *Lottery) GetConfig() Config {
	return l.config
}

func (l *Lottery) getAvailableParticipants(nPrizeIndex int) []Participant {
	participants := l.participants

	// Remove winners
	for _, IDs := range l.winners {
		for _, ID := range IDs {
			delete(participants, ID)
		}
	}

	// Remove blacklists
	for _, blacklist := range l.config.Blacklists {
		if blacklist.MaxPrizeIndex < nPrizeIndex {
			for _, ID := range blacklist.IDs {
				delete(participants, ID)
			}
		}
	}

	return participantsMapToSlice(participants)
}
