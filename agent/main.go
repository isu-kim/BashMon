package main

import (
	"C"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cilium/ebpf/rlimit"
)

// $BPF_CLANG and $BPF_CFLAGS are set by the Makefile.
//
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS -type event bpf monitoring.c --
var master string
var verbose bool

func main() {
	// Get command line arguments
	flag.StringVar(&master, "master", "", "The master node's address")
	flag.BoolVar(&verbose, "verbose", false, "Whether or not to print details")
	flag.Parse()

	if len(master) == 0 {
		log.Fatalf("Must have master's information, use --master argument")
	}

	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatalf("Could not remove memlock: %v", err)
		return
	}

	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("Could load objects: %v", err)
		return
	}
	defer objs.Close()

	// Create a channel to receive SIGINT signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	go uprobeReadLine(objs, sigChan)

	<-sigChan
	log.Println("Received SIGINT signal. Exiting program.")
}
