package lottery

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/northbright/csvhelper"
)

type Participant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Prize struct {
	No     int    `json:"no"`
	Name   string `json:"name"`
	Amount int    `json:"amount"`
	Desc   string `json:"desc"`
}

type Blacklist struct {
	MinPrizeNo int      `json:"min_prize_no"`
	IDs        []string `json:"ids"`
}

type Lottery struct {
	Name         string                 `json:"name"`
	Prizes       map[int]Prize          `json:"prizes"`
	Blacklists   map[int]Blacklist      `json:"blacklists"`
	participants map[string]Participant `json:"participants"`
	winners      map[int][]Participant  `json:"winners"`
	mutex        *sync.Mutex
	appDataDir   string
	dataFile     string
}

type SaveData struct {
	Lottery     Lottery `json:"lottery"`
	LastUpdated string  `json:"last_updated"`
	Checksum    string  `json:"checksum"`
}

const (
	AppName = "lottery-go"
)

var (
	ErrParticipantsCSV               = fmt.Errorf("incorrect participants CSV")
	ErrPrizeNo                       = fmt.Errorf("incorrect prize no")
	ErrWinnersExistBeforeDraw        = fmt.Errorf("winners exist before draw")
	ErrPrizeAmount                   = fmt.Errorf("incorrect prize amount")
	ErrNoAvailableParticipants       = fmt.Errorf("no available participants")
	ErrNoOriginalWinnersBeforeRedraw = fmt.Errorf("no original winners before redraw")
	ErrRevokedWinnerNotMatch         = fmt.Errorf("revoked winner does not match")
	ErrChecksum                      = fmt.Errorf("incorrect checksum")
)

func New(name string) *Lottery {
	l := &Lottery{
		name,
		make(map[int]Prize),
		make(map[int]Blacklist),
		make(map[string]Participant),
		make(map[int][]Participant),
		&sync.Mutex{},
		"",
		"",
	}

	dir, err := l.createAppDataDir()
	if err != nil {
		log.Printf("createAppDataDir() error: %v", err)
		return nil
	}
	l.appDataDir = dir

	l.dataFile = l.makeDataFileName()

	return l
}

func (l *Lottery) SetPrize(no int, name string, amount int, desc string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	prize := Prize{no, name, amount, desc}
	l.Prizes[no] = prize
}

func (l *Lottery) SetBlacklist(minPrizeNo int, IDs []string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	blacklist := Blacklist{minPrizeNo, IDs}
	l.Blacklists[minPrizeNo] = blacklist
}

func (l *Lottery) loadParticipants(file string) error {
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

func (l *Lottery) getAvailableParticipants(nPrizeNo int) []Participant {
	participants := l.participants

	// Remove winners
	for _, winners := range l.winners {
		for _, winner := range winners {
			delete(participants, winner.ID)
		}
	}

	// Remove blacklists
	for _, blacklist := range l.Blacklists {
		if blacklist.MinPrizeNo > nPrizeNo {
			for _, ID := range blacklist.IDs {
				delete(participants, ID)
			}
		}
	}

	return participantMapToSlice(participants)
}

func (l *Lottery) GetAvailableParticipants(nPrizeNo int) []Participant {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.getAvailableParticipants(nPrizeNo)
}

func (l *Lottery) GetWinners(nPrizeNo int) []Participant {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, ok := l.winners[nPrizeNo]; !ok {
		return []Participant{}
	}

	return l.winners[nPrizeNo]
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

func draw(prizeAmount int, participants []Participant) []Participant {
	var winners []Participant

	if prizeAmount <= 0 || len(participants) <= 0 {
		return winners
	}

	// Check prize amount.
	amount := prizeAmount
	// If participants amount < prize amount,
	// use participants amount as the new prize amount.
	if len(participants) < prizeAmount {
		amount = len(participants)
	}

	for i := 0; i < amount; i++ {
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(participants))
		winners = append(winners, participants[index])
		participants = removeParticipant(participants, index)
	}

	return winners
}

func (l *Lottery) Draw(nPrizeNo int) ([]Participant, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var winners []Participant

	if nPrizeNo < 0 || nPrizeNo >= len(l.Prizes) {
		return winners, ErrPrizeNo
	}

	amount := l.Prizes[nPrizeNo].Amount
	if amount < 1 {
		return winners, ErrPrizeAmount
	}

	if _, ok := l.winners[nPrizeNo]; ok {
		return winners, ErrWinnersExistBeforeDraw
	}

	participants := l.getAvailableParticipants(nPrizeNo)
	if len(participants) == 0 {
		return winners, ErrNoAvailableParticipants
	}

	winners = draw(amount, participants)

	l.winners[nPrizeNo] = winners
	return winners, nil
}

func (l *Lottery) Redraw(nPrizeNo int, revokedWinners []Participant) ([]Participant, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var winners []Participant

	if nPrizeNo < 0 || nPrizeNo >= len(l.Prizes) {
		return winners, ErrPrizeNo
	}

	amount := l.Prizes[nPrizeNo].Amount
	if amount < 1 {
		return winners, ErrPrizeAmount
	}

	if _, ok := l.winners[nPrizeNo]; !ok {
		return winners, ErrNoOriginalWinnersBeforeRedraw
	}

	// Remove original winners for the prize before re-draw.
	originalWinnerMap := participantSliceToMap(l.winners[nPrizeNo])

	for _, revokedWinner := range revokedWinners {
		if _, ok := originalWinnerMap[revokedWinner.ID]; !ok {
			return winners, ErrRevokedWinnerNotMatch
		}
		delete(originalWinnerMap, revokedWinner.ID)
	}

	l.winners[nPrizeNo] = participantMapToSlice(originalWinnerMap)

	participants := l.getAvailableParticipants(nPrizeNo)
	if len(participants) == 0 {
		return winners, ErrNoAvailableParticipants
	}

	// Prize amount of redraw = amount of revoked winners.
	amount = len(revokedWinners)

	// Get new winners.
	winners = draw(amount, participants)

	// Append new winners and original winners.
	l.winners[nPrizeNo] = append(l.winners[nPrizeNo], winners...)
	return winners, nil
}

func (l *Lottery) GetAllWinners() map[int][]Participant {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.winners
}

func (l *Lottery) ClearWinners(nPrizeNo int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Clear the winner slice.
	l.winners[nPrizeNo] = []Participant{}
}

func (l *Lottery) ClearAllWinners() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.winners = make(map[int][]Participant)
}

func (l *Lottery) makeDataFileName() string {
	h := md5.New()
	f := fmt.Sprintf("%X.json", h.Sum([]byte(l.Name)))
	return path.Join(l.appDataDir, f)
}

func (l *Lottery) DataFileExists() bool {
	return false
}

func (l *Lottery) createAppDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := path.Join(homeDir, AppName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func computeWinnersHash(winners [][]Participant) []byte {
	h := md5.New()

	for nPrizeNo, winnersOfPrize := range winners {
		s := strconv.FormatInt(int64(nPrizeNo), 10)
		h.Write([]byte(s))
		for _, winner := range winnersOfPrize {
			h.Write([]byte(winner.ID))
			h.Write([]byte(winner.Name))
		}
	}

	return h.Sum(nil)
}

func (l *Lottery) Save() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	winners := winnerMapToSlice(l.winners)
	tm := time.Now()

	data := SaveData{
		l.Name,
		winners,
		fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
			tm.Year(),
			tm.Month(),
			tm.Day(),
			tm.Hour(),
			tm.Minute(),
			tm.Second(),
		),
		fmt.Sprintf("%X", computeWinnersHash(winners)),
	}

	buf, err := json.MarshalIndent(&data, "", "    ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(l.dataFile, buf, 0644); err != nil {
		return err
	}

	return nil
}

func (l *Lottery) Load() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	buf, err := ioutil.ReadFile(l.dataFile)
	if err != nil {
		return err
	}

	data := SaveData{}
	if err := json.Unmarshal(buf, &data); err != nil {
		return err
	}

	checksum := computeWinnersHash(data.Winners)
	if fmt.Sprintf("%X", checksum) != data.Checksum {
		return ErrChecksum
	}

	// Load winners.
	l.winners = winnerSliceToMap(data.Winners)

	return nil
}
