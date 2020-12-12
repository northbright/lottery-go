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
)

func getLotterySettings(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("getLotterySettings\n"))
}

// GetCurrentExecDir gets the current executable path.
// You may find more path helper functions in:
// https://github.com/northbright/pathhelper
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
	l := lottery.New(config.LotteryName)

	// Load participants.
	if err := l.LoadParticipantsCSVFile(participantsCSV); err != nil {
		log.Printf("load participants CSV error: %v", err)
		return
	}
	log.Printf("load participants CSV successfully")
	log.Printf("participants: %v", l.Participants)

	// Load prizes.
	if err := l.LoadPrizesCSVFile(prizesCSV); err != nil {
		log.Printf("load prizes CSV error: %v", err)
		return
	}
	log.Printf("load prizes CSV successfully")
	log.Printf("prizes: %v", l.Prizes)

	// Load blacklists.
	if err := l.LoadBlacklistsJSONFile(blacklistsJSON); err != nil {
		log.Printf("load blacklists JSON error: %v", err)
		return
	}
	log.Printf("load blacklists JSON successfully")
	log.Printf("blacklists: %v", l.Blacklists)

	// Serve Static Files
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(staticFolderPath))))

	// Get lottery settings
	http.HandleFunc("/lottery_settings", getLotterySettings)

	err = http.ListenAndServe(config.Addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
