package mon

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

// ObserveLabel is how we detect containers we want to monitor
const ObserveLabel string = "mon.observe=1"

// CheckHealthLabel is how we detect containers we want to restart based on health
const CheckHealthLabel string = "mon.checks.health=1"

// CheckCleanupLabel is how we detect containers we want to remove if they exit cleanly
const CheckCleanupLabel string = "mon.checks.cleanup=1"

// CleanupExitCodeLabelKey is the label key in which the expected exit code can be overriden
const CleanupExitCodeLabelKey string = "mon.checks.cleanup.code"

// HealthRestartLabelKey is the label key in which the expected restart timeout can be overriden
const HealthRestartLabelKey string = "mon.checks.health.timeout"

// DefaultCleanupExitCode is the default exit code we expect for cleanup
const DefaultCleanupExitCode int = 0

// DefaultRestartTimeoutMs is the default timeout for restarting
const DefaultRestartTimeoutMs int64 = 10 * 1000

// Unhealthy is the literal "Unhealthy"
const Unhealthy string = "Unhealthy"

// Monitor is the core application controller, to monitor and act on containers
type Monitor struct {
	ContainerPrefix string
	Dockerd         DockerAPI
}

func (m *Monitor) handleContainerHealth() {
	conts, err := m.Dockerd.ExecuteListQuery(map[string]string{
		"label": fmt.Sprintf("%s,%s", ObserveLabel, CheckHealthLabel),
	})

	if err != nil {
		log.Printf("ExecuteListQuery failed: %v\n", err)
		return
	}

	for _, cont := range conts {
		//if we have a prefix value, and cont doesn't satisfy it, move along
		if len(m.ContainerPrefix) > 0 && !namesContainPrefix(cont.Names, m.ContainerPrefix) {
			continue
		}

		state, err := parseState(cont.State)
		if err != nil {
			log.Printf("parseState failed: %v\n", err)
		}

		expectedRestartTimeoutMs := DefaultRestartTimeoutMs
		if restartMs, ok := cont.Labels[HealthRestartLabelKey]; ok {
			if i, err := strconv.Atoi(restartMs); err == nil {
				expectedRestartTimeoutMs = int64(i)
			}
		}

		// it's unhealthy and running, restart it
		if (state.Health != nil && state.Health.Status == Unhealthy) && state.Running {
			log.Printf("Found unhealthy running container: %v (%v)\n", cont.ID, cont.Names[0])
			if err := m.Dockerd.Restart(expectedRestartTimeoutMs, cont); err != nil {
				log.Printf("Failed to restart unhealthy container %v (%v): %v\n", cont.ID, cont.Names[0], err)
			}
		}
	}
}

func (m *Monitor) handleContainerCleanup() {
	conts, err := m.Dockerd.ExecuteListQuery(map[string]string{
		"label": fmt.Sprintf("%s,%s", ObserveLabel, CheckCleanupLabel),
	})

	if err != nil {
		log.Printf("ExecuteListQuery failed: %v\n", err)
		return
	}

	for _, cont := range conts {
		//if we have a prefix value, and cont doesn't satisfy it, move along
		if len(m.ContainerPrefix) > 0 && !namesContainPrefix(cont.Names, m.ContainerPrefix) {
			continue
		}

		state, err := parseState(cont.State)
		if err != nil {
			log.Printf("parseState failed: %v\n", err)
		}

		expectedExitCode := DefaultCleanupExitCode
		if errCode, ok := cont.Labels[CleanupExitCodeLabelKey]; ok {
			if i, err := strconv.Atoi(errCode); err == nil {
				expectedExitCode = i
			}
		}

		// it's not running, nor restarting, and exit code is as expected
		if !state.Running && !state.Restarting && state.ExitCode == expectedExitCode {
			log.Printf("Found container to cleanup: %v (%v)\n", cont.ID, cont.Names[0])
			if err := m.Dockerd.Remove(cont); err != nil {
				log.Printf("Failed to remove container %v (%v): %v\n", cont.ID, cont.Names[0], err)
			}
		}
	}
}

// Poll checks the dockerd system and executes operations as needed
func (m *Monitor) Poll(t time.Time) {
	log.Printf("CheckStart for %v\n", t)
	m.handleContainerHealth()
	m.handleContainerCleanup()
	log.Printf("CheckEnd for %v\n", t)
}
