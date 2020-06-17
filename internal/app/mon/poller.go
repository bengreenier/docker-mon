package mon

import (
	"errors"
	"time"
)

// PollHandler provides a poll method
type PollHandler interface {
	Poll(time.Time)
}

// Poller polls the docker socket at an interval, to find containers
type Poller struct {
	IntervalMs     int64
	Handler        PollHandler
	tickerComplete chan bool
	ticker         *time.Ticker
	running        bool
}

// Start begins polling
func (p *Poller) Start() error {
	if p.running {
		return errors.New("Already running")
	}

	p.tickerComplete = make(chan bool)
	p.ticker = time.NewTicker(time.Duration(p.IntervalMs) * time.Millisecond)
	p.running = true

	// run the ticker
	go func() {
		for {
			select {
			case <-p.tickerComplete:
				return
			case t := <-p.ticker.C:
				p.Handler.Poll(t)
			}
		}
	}()

	return nil
}

// Stop ends polling
func (p *Poller) Stop() error {
	if !p.running {
		return errors.New("Not running")
	}

	p.tickerComplete <- true
	p.ticker.Stop()
	p.running = false

	return nil
}
