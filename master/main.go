package main

import (
	"log"
	"net/http"
)

var dh dbHandler

func main() {
	dh = dbHandler{}
	err := dh.initDB()
	if err != nil {
		log.Printf("Could not initialize DB: %v\n", err)
		return
	}

	http.HandleFunc("/event_send", handleEvent)
	log.Fatal(http.ListenAndServe(":9000", nil))
}
