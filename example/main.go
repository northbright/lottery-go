package main

import (
	//"fmt"
	"log"
	"net/http"

	"github.com/northbright/lottery-go/lottery"
)

func main() {
	drawing := &lottery.Drawing{}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		drawing.Handler(w, r)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
