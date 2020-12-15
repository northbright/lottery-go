package main

import (
	"encoding/json"
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

func getLotterySettings(w http.ResponseWriter, r *http.Request) {

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

	// Serve Static Files
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(staticFolderPath))))

	// Get lottery settings
	http.HandleFunc("/lottery_settings", getLotterySettings)

	err = http.ListenAndServe(config.Addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
