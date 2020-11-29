package lottery

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

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
	Name       string      `json:"name"`
	Prizes     []Prize     `json:"prizes"`
	Blacklists []Blacklist `json:"blacklists"`
}

type Lottery struct {
	config       Config
	participants map[string]Participant
	winners      map[int][]Participant
	mutex        *sync.Mutex
}

var (
	ErrParticipantsCSV               = fmt.Errorf("incorrect participants CSV")
	ErrPrizeIndex                    = fmt.Errorf("incorrect prize index")
	ErrWinnersExistBeforeDraw        = fmt.Errorf("winners exist before draw")
	ErrPrizeNum                      = fmt.Errorf("incorrect prize num")
	ErrNoAvailableParticipants       = fmt.Errorf("no available participants")
	ErrNoOriginalWinnersBeforeRedraw = fmt.Errorf("no original winners before redraw")
	ErrRevokedWinnerNotMatch         = fmt.Errorf("revoked winner does not match")
)

func New() *Lottery {
	l := &Lottery{
		Config{},
		make(map[string]Participant),
		make(map[int][]Participant),
		&sync.Mutex{},
	}

	return l
}

func (l *Lottery) LoadParticipants(file string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

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

func participantMapToSlice(m map[string]Participant) []Participant {
	var participants []Participant

	for _, p := range m {
		participants = append(participants, p)
	}

	return participants
}

func participantSliceToMap(s []Participant) map[string]Participant {
	m := make(map[string]Participant)

	for _, p := range s {
		m[p.ID] = p
	}

	return m
}

func (l *Lottery) GetParticipants() []Participant {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return participantMapToSlice(l.participants)
}

func (l *Lottery) LoadConfig(file string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, &l.config)
}

func (l *Lottery) GetConfig() Config {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.config
}

func (l *Lottery) getAvailableParticipants(nPrizeIndex int) []Participant {
	participants := l.participants

	// Remove winners
	for _, winners := range l.winners {
		for _, winner := range winners {
			delete(participants, winner.ID)
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

	return participantMapToSlice(participants)
}

func (l *Lottery) GetAvailableParticipants(nPrizeIndex int) []Participant {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.getAvailableParticipants(nPrizeIndex)
}

func (l *Lottery) GetWinners(nPrizeIndex int) []Participant {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, ok := l.winners[nPrizeIndex]; !ok {
		return []Participant{}
	}

	return l.winners[nPrizeIndex]
}

func removeParticipant(s []Participant, i int) []Participant {
	l := len(s)
	if l <= 0 {
		return s
	}

	if i < 0 || i > l-1 {
		return s
	}

	s[i] = s[l-1]
	return s[:l-1]
}

func draw(prizeNum int, participants []Participant) []Participant {
	var winners []Participant

	if prizeNum <= 0 || len(participants) <= 0 {
		return winners
	}

	// Check prize num.
	num := prizeNum
	// If participants num < prize num,
	// use participants num as the new prize num.
	if len(participants) < prizeNum {
		num = len(participants)
	}

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(participants))
		winners = append(winners, participants[index])
		participants = removeParticipant(participants, index)
	}

	return winners
}

func (l *Lottery) Draw(nPrizeIndex int) ([]Participant, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var winners []Participant

	if nPrizeIndex < 0 || nPrizeIndex >= len(l.config.Prizes) {
		return winners, ErrPrizeIndex
	}

	num := l.config.Prizes[nPrizeIndex].Num
	if num < 1 {
		return winners, ErrPrizeNum
	}

	if _, ok := l.winners[nPrizeIndex]; ok {
		return winners, ErrWinnersExistBeforeDraw
	}

	participants := l.getAvailableParticipants(nPrizeIndex)
	if len(participants) == 0 {
		return winners, ErrNoAvailableParticipants
	}

	winners = draw(num, participants)

	l.winners[nPrizeIndex] = winners
	return winners, nil
}

func (l *Lottery) Redraw(nPrizeIndex int, revokedWinners []Participant) ([]Participant, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var winners []Participant

	if nPrizeIndex < 0 || nPrizeIndex >= len(l.config.Prizes) {
		return winners, ErrPrizeIndex
	}

	num := l.config.Prizes[nPrizeIndex].Num
	if num < 1 {
		return winners, ErrPrizeNum
	}

	if _, ok := l.winners[nPrizeIndex]; !ok {
		return winners, ErrNoOriginalWinnersBeforeRedraw
	}

	// Remove original winners for the prize before re-draw.
	originalWinnerMap := participantSliceToMap(l.winners[nPrizeIndex])

	for _, revokedWinner := range revokedWinners {
		if _, ok := originalWinnerMap[revokedWinner.ID]; !ok {
			return winners, ErrRevokedWinnerNotMatch
		}
		delete(originalWinnerMap, revokedWinner.ID)
	}

	l.winners[nPrizeIndex] = participantMapToSlice(originalWinnerMap)

	participants := l.getAvailableParticipants(nPrizeIndex)
	if len(participants) == 0 {
		return winners, ErrNoAvailableParticipants
	}

	// Prize num of redraw = num of revoked winners.
	num = len(revokedWinners)

	// Get new winners.
	winners = draw(num, participants)

	// Append new winners and original winners.
	l.winners[nPrizeIndex] = append(l.winners[nPrizeIndex], winners...)
	return winners, nil
}
