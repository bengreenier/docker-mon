package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/bengreenier/docker-mon/internal/app/mon"
)

var control = flag.String("control", "unix:///var/run/docker.sock", "Docker control socket")
var prefix = flag.String("prefix", "", "Docker container prefix to limit observation to")
var interval = flag.Int64("interval", 5000, "Interval to poll at (in ms)")
var retries = flag.Int64("retries", 10, "Max retry count for failed docker commands")
var quiet = flag.Bool("quiet", false, "Only log when action is taken")

func main() {
	flag.Parse()

	// TODO(bengreenier): flip this so cli overrides env
	if s, ok := envStr("MON_CONTROL"); ok {
		*control = s
	}
	if s, ok := envStr("MON_PREFIX"); ok {
		*prefix = s
	}
	if i, ok := envInt64("MON_INTERVAL"); ok {
		*interval = i
	}
	if i, ok := envInt64("MON_RETRIES"); ok {
		*retries = i
	}
	if b, ok := envBool("MON_QUIET"); ok {
		*quiet = b
	}

	log.Printf("control: '%s', prefix: '%s', interval: '%v', retries: %v, quiet: %v\n", *control, *prefix, *interval, *retries, *quiet)

	monitor := mon.Monitor{
		Quiet: *quiet,
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

func envStr(key string) (string, bool) {
	return os.LookupEnv(key)
}

func envInt64(key string) (int64, bool) {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return int64(i), true
		}
	}

	return 0, false
}

func envBool(key string) (bool, bool) {
	if val, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(val); err == nil {
			return b, true
		}
	}

	return false, false
}
