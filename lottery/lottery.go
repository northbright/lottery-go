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
	"sort"
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
	Name string `json:"name"`
	Num  int    `json:"num"`
	Desc string `json:"desc"`
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

type SaveData struct {
	Name        string          `json:"name"`
	Winners     [][]Participant `json:"winners"`
	LastUpdated string          `json:"last_updated"`
	Checksum    string          `json:"checksum"`
}

type Lottery struct {
	config       Config
	participants map[string]Participant
	winners      map[int][]Participant
	mutex        *sync.Mutex
	appDataDir   string
	dataFile     string
}

const (
	AppName = "lottery-go"
)

var (
	ErrParticipantsCSV               = fmt.Errorf("incorrect participants CSV")
	ErrPrizeIndex                    = fmt.Errorf("incorrect prize index")
	ErrWinnersExistBeforeDraw        = fmt.Errorf("winners exist before draw")
	ErrPrizeNum                      = fmt.Errorf("incorrect prize num")
	ErrNoAvailableParticipants       = fmt.Errorf("no available participants")
	ErrNoOriginalWinnersBeforeRedraw = fmt.Errorf("no original winners before redraw")
	ErrRevokedWinnerNotMatch         = fmt.Errorf("revoked winner does not match")
	ErrChecksum                      = fmt.Errorf("incorrect checksum")
)

func New(csvFile, configFile string) *Lottery {
	l := &Lottery{
		Config{},
		make(map[string]Participant),
		make(map[int][]Participant),
		&sync.Mutex{},
		"",
		"",
	}

	if err := l.loadParticipants(csvFile); err != nil {
		log.Printf("loadParticipants() error: %v", err)
		return nil
	}

	if err := l.loadConfig(configFile); err != nil {
		log.Printf("loadConfig() error: %v", err)
		return nil
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

func (l *Lottery) loadConfig(file string) error {
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

func winnerMapToSlice(m map[int][]Participant) [][]Participant {
	var (
		arr     []int
		winners [][]Participant
	)

	// Sort map by key(prize index).
	for k, _ := range m {
		arr = append(arr, k)
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})

	for _, prizeIndex := range arr {
		winners = append(winners, m[prizeIndex])
	}

	return winners
}

func winnerSliceToMap(s [][]Participant) map[int][]Participant {
	m := make(map[int][]Participant)

	for nPrizeIndex, winnersOfPrize := range s {
		m[nPrizeIndex] = winnersOfPrize
	}

	return m
}

func computeWinnersHash(winners [][]Participant) []byte {
	h := md5.New()

	for nPrizeIndex, winnersOfPrize := range winners {
		s := strconv.FormatInt(int64(nPrizeIndex), 10)
		h.Write([]byte(s))
		for _, winner := range winnersOfPrize {
			h.Write([]byte(winner.ID))
			h.Write([]byte(winner.Name))
		}
	}

	return h.Sum(nil)
}

func (l *Lottery) GetAllWinners() [][]Participant {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return winnerMapToSlice(l.winners)
}

func (l *Lottery) makeDataFileName() string {
	h := md5.New()
	f := fmt.Sprintf("%X.json", h.Sum([]byte(l.config.Name)))
	return path.Join(l.appDataDir, f)
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

func (l *Lottery) Save() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	winners := winnerMapToSlice(l.winners)
	tm := time.Now()

	data := SaveData{
		l.config.Name,
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
