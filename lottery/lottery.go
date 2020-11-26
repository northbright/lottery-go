package lottery

import (
	//"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
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

func (l *Lottery) GetParticipants() []Participant {
	var participants []Participant

	// Sort participants by IDs
	for _, p := range l.participants {
		participants = append(participants, p)
	}

	sort.Slice(participants, func(i, j int) bool {
		return participants[i].ID < participants[j].ID
	})

	return participants
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
