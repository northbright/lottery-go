package lottery

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
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
	winners      map[int][]Participant
	mutex        *sync.Mutex
}

var (
	ErrParticipantsCSV         = fmt.Errorf("incorrect participants CSV")
	ErrPrizeIndex              = fmt.Errorf("incorrect prize index")
	ErrWinnersExistBeforeDraw  = fmt.Errorf("winners exist before draw")
	ErrPrizeNum                = fmt.Errorf("incorrect prize num")
	ErrNoAvailableParticipants = fmt.Errorf("no available participants")
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
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return participantsMapToSlice(l.participants)
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

	return participantsMapToSlice(participants)
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

	for i := 0; i < num; i++ {
		rand.Seed(time.Now().UnixNano())
		index := rand.Intn(len(participants))
		winners = append(winners, participants[index])
		participants = removeParticipant(participants, index)
	}

	l.winners[nPrizeIndex] = winners
	return winners, nil
}
