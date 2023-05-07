package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

// generateEvent generates eventInfo from a given context.
func generateEvent(ctx context) eventInfo {
	// Generate event information from context.
	ret := eventInfo{
		Hostname: getHostname(),
		Pid:      ctx.pid,
		Ppid:     ctx.ppid,
		PpName:   ctx.parentProcName,
		Uid:      ctx.uid,
		Username: getUsername(int(ctx.uid)),
		Command:  ctx.line,
	}

	// Retrieve the container's name
	containerName := getProcInfo(int(ctx.pid))
	if containerName != "" { // This was a container.
		ret.Container = containerName
		ret.IsContainer = true
	} else { // This was not a container.
		ret.Container = ""
		ret.IsContainer = false
	}

	return ret
}

// sendEvent sends current bash event to the API host.
func sendEvent(ctx context, url string) error {
	payload := generateEvent(ctx)
	// Marshal the payload into a JSON string
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create a new HTTP request with the payload
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	// Set the Content-Type header to application/json
	req.Header.Set("Content-Type", "application/json")

	// Send the request and get the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response code.
	if resp.StatusCode != 200 {
		return errors.New("Could server responded with " + resp.Status)
	}

	return nil
}
