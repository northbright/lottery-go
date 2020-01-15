package lottery

import (
	"fmt"
	"log"
	"net/http"
)

type Drawing struct {
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
