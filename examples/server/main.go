package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/northbright/lottery-go/lottery"
)

type Config struct {
	// Addr is the HTTP server address. Default address is ":8080".
	Addr string `json:"addr"`
	// LotteryName is the lottery name.
	LotteryName string `json:"lottery_name"`
}

var (
	serverRoot       string // Absolute path of server root.
	staticFolderPath string // Absolute path of static file folder.
	participantsCSV  string
	prizesCSV        string
	blacklistsJSON   string
	lott             *lottery.Lottery
)

// prizes returns the prizes.
func prizes(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Success bool            `json:"success"`
		ErrMsg  string          `json:"err_msg,omitempty"`
		Prizes  []lottery.Prize `json:"prizes"`
	}

	var (
		errMsg string
		prizes []lottery.Prize
	)

	defer func() {
		resp := Response{}

		if errMsg == "" {
			resp.Success = true
		} else {
			resp.Success = false
			resp.ErrMsg = errMsg
			log.Printf("prizes(): error: %v", errMsg)
		}

		resp.Prizes = prizes

		w.Header().Set("Content-Type", "application/json")

		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		if err := enc.Encode(&resp); err != nil {
			log.Printf("prizes() encode JSON error: %v", err)
			return
		}
	}()

	if r.Method != "GET" {
		errMsg = fmt.Sprintf("prizes(): HTTP method is NOT GET(%v)", r.Method)
		return
	}

	prizes = lott.Prizes(true)
}

// availableParticipants returns the available participants for given prize no.
func availableParticipants(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		PrizeNo int `json:"prize_no"`
	}

	type Response struct {
		Success               bool                  `json:"success"`
		ErrMsg                string                `json:"err_msg,omitempty"`
		PrizeNo               int                   `json:"prize_no"`
		AvailableParticipants []lottery.Participant `json:"available_participants"`
	}

	var (
		errMsg                string
		req                   Request
		availableParticipants []lottery.Participant
	)

	defer func() {
		resp := Response{}

		if errMsg == "" {
			resp.Success = true
		} else {
			resp.Success = false
			resp.ErrMsg = errMsg
			log.Printf("availableParticipants(): error: %v", errMsg)
		}

		resp.PrizeNo = req.PrizeNo
		resp.AvailableParticipants = availableParticipants

		w.Header().Set("Content-Type", "application/json")

		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		if err := enc.Encode(&resp); err != nil {
			log.Printf("availableParticipants() encode JSON error: %v", err)
			return
		}
	}()

	if r.Method != "POST" {
		errMsg = fmt.Sprintf("revoke(): HTTP method is NOT POST(%v)", r.Method)
		return
	}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		errMsg = fmt.Sprintf("revoke(): decode JSON error: %v", err)
		return
	}

	availableParticipants = lott.AvailableParticipants(req.PrizeNo)
}

// winners returns the winners of a prize.
func winners(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		PrizeNo int `json:"prize_no"`
	}

	type Response struct {
		Success bool                  `json:"success"`
		ErrMsg  string                `json:"err_msg,omitempty"`
		PrizeNo int                   `json:"prize_no"`
		Winners []lottery.Participant `json:"winners"`
	}

	var (
		errMsg  string
		req     Request
		winners []lottery.Participant
	)

	defer func() {
		resp := Response{}

		if errMsg == "" {
			resp.Success = true
		} else {
			resp.Success = false
			resp.ErrMsg = errMsg
			log.Printf("winners(): error: %v", errMsg)
		}

		resp.PrizeNo = req.PrizeNo
		resp.Winners = winners

		w.Header().Set("Content-Type", "application/json")

		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		if err := enc.Encode(&resp); err != nil {
			log.Printf("winners() encode JSON error: %v", err)
			return
		}
	}()
	if r.Method != "POST" {
		errMsg = fmt.Sprintf("winners(): HTTP method is NOT POST(%v)", r.Method)
		return
	}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		errMsg = fmt.Sprintf("winners(): decode JSON error: %v", err)
		return
	}

	winners = lott.Winners(req.PrizeNo)
}

// draw draws a prize and returns the winners.
func draw(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		PrizeNo int `json:"prize_no"`
	}

	type Response struct {
		Success bool                  `json:"success"`
		ErrMsg  string                `json:"err_msg,omitempty"`
		PrizeNo int                   `json:"prize_no"`
		Winners []lottery.Participant `json:"winners"`
	}

	var (
		errMsg  string
		req     Request
		winners []lottery.Participant
	)

	defer func() {
		resp := Response{}

		if errMsg == "" {
			resp.Success = true
		} else {
			resp.Success = false
			resp.ErrMsg = errMsg
			log.Printf("draw(): error: %v", errMsg)
		}

		resp.PrizeNo = req.PrizeNo
		resp.Winners = winners

		w.Header().Set("Content-Type", "application/json")

		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		if err := enc.Encode(&resp); err != nil {
			log.Printf("draw() encode JSON error: %v", err)
			return
		}
	}()

	if r.Method != "POST" {
		errMsg = fmt.Sprintf("draw(): HTTP method is NOT POST(%v)", r.Method)
		return
	}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		errMsg = fmt.Sprintf("draw(): decode JSON error: %v", err)
		return
	}

	winners, err := lott.Draw(req.PrizeNo)
	if err != nil {
		errMsg = fmt.Sprintf("draw(): Draw() error: %v", err)
		return
	}
}

// revoke revokes the winners of given prize no.
func revoke(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		PrizeNo        int                   `json:"prize_no"`
		RevokedWinners []lottery.Participant `json:"revoked_winners"`
	}

	type Response struct {
		Success        bool                  `json:"success"`
		ErrMsg         string                `json:"err_msg,omitempty"`
		PrizeNo        int                   `json:"prize_no"`
		RevokedWinners []lottery.Participant `json:"revoked_winners"`
	}

	var (
		errMsg string
		req    Request
	)

	defer func() {
		resp := Response{}

		if errMsg == "" {
			resp.Success = true
		} else {
			resp.Success = false
			resp.ErrMsg = errMsg
			log.Printf("revoke(): error: %v", errMsg)
		}

		resp.PrizeNo = req.PrizeNo
		resp.RevokedWinners = req.RevokedWinners

		w.Header().Set("Content-Type", "application/json")

		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		if err := enc.Encode(&resp); err != nil {
			log.Printf("revoke() encode JSON error: %v", err)
			return
		}
	}()

	if r.Method != "POST" {
		errMsg = fmt.Sprintf("revoke(): HTTP method is NOT POST(%v)", r.Method)
		return
	}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		errMsg = fmt.Sprintf("revoke(): decode JSON error: %v", err)
		return
	}

	if err := lott.Revoke(req.PrizeNo, req.RevokedWinners); err != nil {
		errMsg = fmt.Sprintf("revoke(): Revoke() error: %v", err)
		return
	}
}

// redraw re-draws a prize with given prize no and amount.
func redraw(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		PrizeNo int `json:"prize_no"`
		Amount  int `json:"amount"`
	}

	type Response struct {
		Success bool                  `json:"success"`
		ErrMsg  string                `json:"err_msg,omitempty"`
		PrizeNo int                   `json:"prize_no"`
		Amount  int                   `json:"amount"`
		Winners []lottery.Participant `json:"winners"`
	}

	var (
		errMsg  string
		req     Request
		winners []lottery.Participant
	)

	defer func() {
		resp := Response{}

		if errMsg == "" {
			resp.Success = true
		} else {
			resp.Success = false
			resp.ErrMsg = errMsg
			log.Printf("redraw(): error: %v", errMsg)
		}

		resp.PrizeNo = req.PrizeNo
		resp.Amount = req.Amount
		resp.Winners = winners

		w.Header().Set("Content-Type", "application/json")

		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		if err := enc.Encode(&resp); err != nil {
			log.Printf("redraw() encode JSON error: %v", err)
			return
		}
	}()

	if r.Method != "POST" {
		errMsg = fmt.Sprintf("redraw(): HTTP method is NOT POST(%v)", r.Method)
		return
	}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		errMsg = fmt.Sprintf("redraw(): decode JSON error: %v", err)
		return
	}

	winners, err := lott.Redraw(req.PrizeNo, req.Amount)
	if err != nil {
		errMsg = fmt.Sprintf("redraw(): Redraw() error: %v", err)
		return
	}
}

// GetCurrentExecDir gets the current executable path.
func GetCurrentExecDir() (dir string, err error) {
	p, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}

	dir = filepath.Dir(absPath)
	return dir, nil
}

func loadConfig() (Config, error) {
	config := Config{}

	f := path.Join(serverRoot, "settings/config.json")

	buf, err := ioutil.ReadFile(f)
	if err != nil {
		return config, err
	}

	if err := json.Unmarshal(buf, &config); err != nil {
		return config, err
	}
	return config, nil
}

func init() {
	// Get absolute path of server root(current executable).
	serverRoot, _ = GetCurrentExecDir()
	// Get static folder path.
	staticFolderPath = path.Join(serverRoot, "./statics/dist/spa")
	log.Printf("staticFolderPath: %v", staticFolderPath)

	participantsCSV = path.Join(serverRoot, "./settings/participants.csv")
	prizesCSV = path.Join(serverRoot, "./settings/prizes.csv")
	blacklistsJSON = path.Join(serverRoot, "./settings/blacklists.json")
}

func main() {
	// Load config.
	config, err := loadConfig()
	if err != nil {
		log.Printf("loadConfig() error: %v", err)
		return
	}

	log.Printf("load config successfully. config: %v", config)

	// Create a lottery.
	lott = lottery.New(config.LotteryName)

	// Check if data file is already saved.
	if lott.DataFileExists() {
		// The lottery started and saved the data.
		// Load the data and continue.
		log.Printf("saved data file found")
		if err := lott.LoadFromFile(); err != nil {
			log.Printf("load data file error: %v", err)
			return
		}
	} else {
		// 1st run for the lottery.
		// Load participants.
		if err := lott.LoadParticipantsCSVFile(participantsCSV); err != nil {
			log.Printf("load participants CSV error: %v", err)
			return
		}
		log.Printf("load participants CSV successfully")
		log.Printf("participants: %v", lott.Participants())

		// Load prizes.
		if err := lott.LoadPrizesCSVFile(prizesCSV); err != nil {
			log.Printf("load prizes CSV error: %v", err)
			return
		}
		log.Printf("load prizes CSV successfully")
		log.Printf("prizes: %v", lott.Prizes(true))

		// Load blacklists.
		if err := lott.LoadBlacklistsJSONFile(blacklistsJSON); err != nil {
			log.Printf("load blacklists JSON error: %v", err)
			return
		}
		log.Printf("load blacklists JSON successfully")
		log.Printf("blacklists: %v", lott.Blacklists())
	}

	// Serve Static Files.
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(staticFolderPath))))

	// Get prizes.
	http.HandleFunc("/prizes", prizes)

	// Get available participants.
	http.HandleFunc("/available_participants", availableParticipants)

	// Get winners.
	http.HandleFunc("/winners", winners)

	// Draw a prize.
	http.HandleFunc("/draw", draw)

	// Revoke winners.
	http.HandleFunc("/revoke", revoke)

	// Redraw a prize.
	http.HandleFunc("/redraw", redraw)

	err = http.ListenAndServe(config.Addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
