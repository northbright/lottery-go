package lottery

import (
//"encoding/json"
//"fmt"
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
	Config
	Participants []Participant
	Winners      map[int][]Participant
}
