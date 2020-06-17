package mon

import (
	"errors"
	"log"
	"testing"
	"time"

	"github.com/docker/engine/pkg/testutil/assert"
)

type mockHandler struct {
	count int
}

func TestPoll(t *testing.T) {
	handler := mockHandler{}

	poll := Poller{
		IntervalMs: 100,
		Handler:    &handler,
	}

	assert.NilError(t, poll.Start())
	assert.Error(t, poll.Start(), "Already running")
	time.Sleep(2 * time.Second)
	assert.NilError(t, poll.Stop())
	assert.Error(t, poll.Stop(), "Not running")

	if handler.count < 18 || handler.count > 22 {
		log.Printf("handler count: %v", handler.count)
		assert.NotNil(t, errors.New("poller did not successfully"))
	}
}

func (m *mockHandler) Poll(t time.Time) {
	m.count++
}
