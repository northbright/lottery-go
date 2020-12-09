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
	"strings"
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
	Participants map[string]Participant `json:"participants"`
	Winners      map[int][]Participant  `json:"winners"`
	mutex        *sync.Mutex
}

type SaveData struct {
	Lottery     *Lottery `json:"lottery"`
	LastUpdated string   `json:"last_updated"`
	Checksum    string   `json:"checksum"`
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
	AppDataDir                       string
)

func init() {
	dir, err := CreateAppDataDir()
	if err != nil {
		log.Printf("CreateAppDataDir() error: %v", err)
		return
	}
	AppDataDir = dir
}

func New(name string) *Lottery {
	l := &Lottery{
		name,
		make(map[int]Prize),
		make(map[int]Blacklist),
		make(map[string]Participant),
		make(map[int][]Participant),
		&sync.Mutex{},
	}

	return l
}

func Load(name string) (*Lottery, error) {
	l := New(name)
	if err := l.Load(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Lottery) SetPrize(no int, name string, amount int, desc string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	prize := Prize{no, name, amount, desc}
	l.Prizes[no] = prize
}

func (l *Lottery) SetPrizesFromCSV(file string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	rows, err := csvhelper.ReadFile(file)
	if err != nil {
		return err
	}

	l.Prizes = make(map[int]Prize)
	for i := 1; i < len(rows); i++ {
		row := rows[i]

		if len(row) != 4 {
			return ErrParticipantsCSV
		}
		no, err := strconv.Atoi(strings.Trim(row[0], " "))
		if err != nil {
			return err
		}
		name := row[1]
		amount, err := strconv.Atoi(strings.Trim(row[2], " "))
		if err != nil {
			return err
		}
		desc := row[3]

		l.Prizes[no] = Prize{no, name, amount, desc}
	}
	return nil
}

func (l *Lottery) SetBlacklist(minPrizeNo int, IDs []string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	blacklist := Blacklist{minPrizeNo, IDs}
	l.Blacklists[minPrizeNo] = blacklist
}

func (l *Lottery) SetBlacklistsFromJSON(f string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	buf, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, &l.Blacklists)
}

func (l *Lottery) SetParticipantsFromCSV(file string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	rows, err := csvhelper.ReadFile(file)
	if err != nil {
		return err
	}

	l.Participants = make(map[string]Participant)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) != 2 {
			return ErrParticipantsCSV
		}
		ID := row[0]
		name := row[1]
		l.Participants[ID] = Participant{ID, name}
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

	return participantMapToSlice(l.Participants)
}

func copyParticipantMap(m map[string]Participant) map[string]Participant {
	copiedMap := make(map[string]Participant)

	for k, v := range m {
		copiedMap[k] = v
	}

	return copiedMap
}

func (l *Lottery) getAvailableParticipants(nPrizeNo int) []Participant {
	participants := copyParticipantMap(l.Participants)

	// Remove winners
	for _, winners := range l.Winners {
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

	if _, ok := l.Winners[nPrizeNo]; !ok {
		return []Participant{}
	}

	return l.Winners[nPrizeNo]
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

	if _, ok := l.Prizes[nPrizeNo]; !ok {
		return winners, ErrPrizeNo
	}

	amount := l.Prizes[nPrizeNo].Amount
	if amount < 1 {
		return winners, ErrPrizeAmount
	}

	if _, ok := l.Winners[nPrizeNo]; ok {
		return winners, ErrWinnersExistBeforeDraw
	}

	participants := l.getAvailableParticipants(nPrizeNo)
	if len(participants) == 0 {
		return winners, ErrNoAvailableParticipants
	}

	winners = draw(amount, participants)

	l.Winners[nPrizeNo] = winners
	return winners, nil
}

func (l *Lottery) Redraw(nPrizeNo int, revokedWinners []Participant) ([]Participant, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var winners []Participant

	if _, ok := l.Prizes[nPrizeNo]; !ok {
		return winners, ErrPrizeNo
	}

	amount := l.Prizes[nPrizeNo].Amount
	if amount < 1 {
		return winners, ErrPrizeAmount
	}

	if _, ok := l.Winners[nPrizeNo]; !ok {
		return winners, ErrNoOriginalWinnersBeforeRedraw
	}

	// Remove original winners for the prize before re-draw.
	originalWinnerMap := participantSliceToMap(l.Winners[nPrizeNo])

	for _, revokedWinner := range revokedWinners {
		if _, ok := originalWinnerMap[revokedWinner.ID]; !ok {
			return winners, ErrRevokedWinnerNotMatch
		}
		delete(originalWinnerMap, revokedWinner.ID)
	}

	l.Winners[nPrizeNo] = participantMapToSlice(originalWinnerMap)

	participants := l.getAvailableParticipants(nPrizeNo)
	if len(participants) == 0 {
		return winners, ErrNoAvailableParticipants
	}

	// Prize amount of redraw = amount of revoked winners.
	amount = len(revokedWinners)

	// Get new winners.
	winners = draw(amount, participants)

	// Append new winners and original winners.
	l.Winners[nPrizeNo] = append(l.Winners[nPrizeNo], winners...)
	return winners, nil
}

func (l *Lottery) GetAllWinners() map[int][]Participant {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.Winners
}

func (l *Lottery) ClearWinners(nPrizeNo int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Clear the winner slice.
	l.Winners[nPrizeNo] = []Participant{}
}

func (l *Lottery) ClearAllWinners() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.Winners = make(map[int][]Participant)
}

func makeDataFileName(name string) string {
	f := fmt.Sprintf("%X.json", md5.Sum([]byte(name)))
	return path.Join(AppDataDir, f)
}

func (l *Lottery) DataFileExists() bool {
	return false
}

func CreateAppDataDir() (string, error) {
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

func computeWinnersHash(winners map[int][]Participant) []byte {
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

	tm := time.Now()

	data := SaveData{
		l,
		fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
			tm.Year(),
			tm.Month(),
			tm.Day(),
			tm.Hour(),
			tm.Minute(),
			tm.Second(),
		),
		fmt.Sprintf("%X", computeWinnersHash(l.Winners)),
	}

	buf, err := json.MarshalIndent(&data, "", "    ")
	if err != nil {
		return err
	}

	dataFile := makeDataFileName(l.Name)
	if err := ioutil.WriteFile(dataFile, buf, 0644); err != nil {
		return err
	}

	return nil
}

func (l *Lottery) Load() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	dataFile := makeDataFileName(l.Name)
	buf, err := ioutil.ReadFile(dataFile)
	if err != nil {
		return err
	}

	data := SaveData{}
	if err := json.Unmarshal(buf, &data); err != nil {
		return err
	}

	checksum := computeWinnersHash(l.Winners)
	if fmt.Sprintf("%X", checksum) != data.Checksum {
		return ErrChecksum
	}

	l.Prizes = data.Lottery.Prizes
	l.Blacklists = data.Lottery.Blacklists
	l.Participants = data.Lottery.Participants
	l.Winners = data.Lottery.Winners

	return nil
}
