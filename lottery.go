package lottery

import (
	//"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	//"math/rand"
	"path"
	//"sync"
	//"time"

	"github.com/northbright/csvhelper"
	"github.com/northbright/pathhelper"
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
	participants []Participant
	winners      map[int][]Participant
}

func (l *Lottery) LoadParticipants(file string) error {
	currentDir, _ := pathhelper.GetCurrentExecDir()
	file = path.Join(currentDir, file)

	rows, err := csvhelper.ReadFile(file)
	if err != nil {
		return err
	}

	l.participants = []Participant{}
	for _, row := range rows {
		if len(row) != 2 {
			return fmt.Errorf("incorrect participants CSV")
		}
		l.participants = append(l.participants, Participant{row[0], row[1]})
	}
	return nil
}

func (l *Lottery) LoadConfig(file string) error {
	currentDir, _ := pathhelper.GetCurrentExecDir()
	file = path.Join(currentDir, file)

	// Load Conifg.
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, &l.config)
}
