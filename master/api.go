package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type eventInfo struct {
	Hostname    string `json:"hostname"`
	Pid         uint32 `json:"pid"`
	Ppid        uint32 `json:"ppid"`
	PpName      string `json:"ppName"`
	Uid         uint32 `json:"uid"`
	Username    string `json:"username"`
	Command     string `json:"command"`
	Container   string `json:"container"`
	IsContainer bool   `json:"isContainer"`
}

// handleEvent is a handler function for new_event endpoint
func handleEvent(w http.ResponseWriter, r *http.Request) {
	// Decode JSON payload into eventInfo struct
	var info eventInfo
	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("%v", info)
	if k8sEnabled {
		log.Println(getPodFromContainer(info.Container))
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Event received"))
	err = dh.insertEvent(info)
	if err != nil {
		log.Printf("Could not insert into DB: %v\n", err)
	}
}
