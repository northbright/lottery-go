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

// getLottery returns the lottery data.
func getLottery(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	if err := enc.Encode(&lott); err != nil {
		log.Printf("getLottery() encode error: %v", err)
		return
	}
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
			resp.PrizeNo = req.PrizeNo
			resp.Winners = winners
		} else {
			resp.Success = false
			resp.ErrMsg = errMsg
			log.Printf("draw(): error: %v", errMsg)
		}

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

	log.Printf("draw() json.Unmarshal() successfully. req: %v", req)

	winners, err := lott.Draw(req.PrizeNo)
	if err != nil {
		errMsg = fmt.Sprintf("draw(): Draw() error: %v", err)
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
		log.Printf("participants: %v", lott.Participants)

		// Load prizes.
		if err := lott.LoadPrizesCSVFile(prizesCSV); err != nil {
			log.Printf("load prizes CSV error: %v", err)
			return
		}
		log.Printf("load prizes CSV successfully")
		log.Printf("prizes: %v", lott.Prizes)

		// Load blacklists.
		if err := lott.LoadBlacklistsJSONFile(blacklistsJSON); err != nil {
			log.Printf("load blacklists JSON error: %v", err)
			return
		}
		log.Printf("load blacklists JSON successfully")
		log.Printf("blacklists: %v", lott.Blacklists)
	}

	// Serve Static Files.
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(staticFolderPath))))

	// Get lottery data.
	http.HandleFunc("/lottery", getLottery)

	// Draw a prize.
	http.HandleFunc("/draw", draw)

	err = http.ListenAndServe(config.Addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
