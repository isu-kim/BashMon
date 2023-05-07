package main

import (
	"flag"
	"log"
	"net/http"
)

var dh dbHandler
var k8sEnabled bool

func main() {
	flag.BoolVar(&k8sEnabled, "k8s", false, "Whether or not Kubernetes is enabled or not")
	flag.Parse()

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
