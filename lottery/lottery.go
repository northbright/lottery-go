package lottery

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(*http.Request) bool { return true },
	}
)

type Settings struct {
}

type State struct {
}

type DB interface {
	ListIDs() []string
	Load(ID string) (*Drawing, error)
	Save(ID string, d *Drawing) error
	LoadSettings(ID string) (*Settings, error)
	SaveSettings(ID string, s *Settings) error
	LoadState(ID string) (*State, error)
	SaveState(ID string, s *State) error
}

type Drawing struct {
	id       string
	settings Settings
	state    State
	db       *DB
}

func (d *Drawing) Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Drawing Handler\n")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, 256), drawing: d}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
