package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
)

var dh dbHandler
var k8sEnabled bool
var verbose bool
var port int

func main() {
	flag.BoolVar(&k8sEnabled, "k8s", false, "Whether or not Kubernetes is enabled or not")
	flag.BoolVar(&verbose, "verbose", false, "Whether or not to print details")
	flag.IntVar(&port, "port", 9000, "Port to listen on")
	flag.Parse()

	log.Println("BashMon Master")
	log.Printf("Listening on 0.0.0.0:%d\n", port)
	log.Printf("Verbose: %s\n", strconv.FormatBool(verbose))
	log.Printf("Kubernetes Enabled: %s\n", strconv.FormatBool(k8sEnabled))

	dh = dbHandler{}
	err := dh.initDB()
	knownPods = make(map[string]string)

	if err != nil {
		log.Printf("Could not initialize DB: %v\n", err)
		return
	}

	http.HandleFunc("/event_send", handleEvent)
	log.Fatal(http.ListenAndServe(":9000", nil))
}
