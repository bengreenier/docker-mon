package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bengreenier/docker-mon/internal/app/mon"
)

var control = flag.String("control", "unix:///var/run/docker.sock", "Docker control socket")
var prefix = flag.String("prefix", "", "Docker container prefix to limit observation to")
var interval = flag.Int64("interval", 5000, "Interval to poll at (in ms)")
var retries = flag.Int64("retries", 10, "Max retry count for failed docker commands")

func main() {
	flag.Parse()

	log.Printf("control: '%s', prefix: '%s', interval: '%v', retries: %v\n", *control, *prefix, *interval, *retries)

	monitor := mon.Monitor{
		Dockerd: &mon.DockerD{
			ControlAddr: *control,
			// see https://docs.docker.com/engine/api/#api-version-matrix
			TargetVersion:  "1.37",
			CommandRetries: *retries,
		},
		ContainerPrefix: *prefix,
	}

	poll := mon.Poller{
		IntervalMs: *interval,
		Handler:    &monitor,
	}

	// startup errors trigger immediate exit
	if err := poll.Start(); err != nil {
		panic(err)
	}

	WaitForTERM()

	if err := poll.Stop(); err != nil {
		fmt.Printf("error on shutdown: %v\n", err)
	}
}

// WaitForTERM waits for the SIGTERM signal, then returns
func WaitForTERM() {
	sigs := make(chan os.Signal, 1)
	term := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Printf("Caught %v, stopping", sig)
		term <- true
	}()

	<-term
}
