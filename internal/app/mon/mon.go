package mon

import (
	"log"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
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

// RunningState is the literal "running"
const RunningState string = "running"

// ExitedState is the literal "exited"
const ExitedState string = "exited"

// Monitor is the core application controller, to monitor and act on containers
type Monitor struct {
	ContainerPrefix string
	Dockerd         DockerAPI
	Quiet           bool
}

func (m *Monitor) handleContainerHealth() {
	conts, err := m.Dockerd.ExecuteListQuery([]string{
		ObserveLabel,
		CheckHealthLabel,
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

		expectedRestartTimeoutMs := DefaultRestartTimeoutMs
		if restartMs, ok := cont.Labels[HealthRestartLabelKey]; ok {
			if i, err := strconv.Atoi(restartMs); err == nil {
				expectedRestartTimeoutMs = int64(i)
			}
		}

		// if it's running, we might need to restart it - we guard the "expensive" inspect call this way
		if cont.State == RunningState {
			if !m.Quiet {
				log.Printf("Checking container health: %v (%v)\n", cont.ID, cont.Names[0])
			}

			inspect, err := m.Dockerd.Inspect(cont)
			if err != nil {
				log.Printf("Inspect failed: %v\n", err)
				continue
			}

			if inspect.State != nil && inspect.State.Health != nil && inspect.State.Health.Status == types.Unhealthy {
				if !m.Quiet {
					log.Printf("Found unhealthy running container: %v (%v)\n", cont.ID, cont.Names[0])
				}
				if err := m.Dockerd.Restart(expectedRestartTimeoutMs, cont); err != nil {
					log.Printf("Failed to restart unhealthy container %v (%v): %v\n", cont.ID, cont.Names[0], err)
				} else {
					log.Printf("Container restarted: %v (%v)\n", cont.ID, cont.Names[0])
				}
			}
		}
	}
}

func (m *Monitor) handleContainerCleanup() {
	conts, err := m.Dockerd.ExecuteListQuery([]string{
		ObserveLabel,
		CheckCleanupLabel,
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

		expectedExitCode := DefaultCleanupExitCode
		if errCode, ok := cont.Labels[CleanupExitCodeLabelKey]; ok {
			if i, err := strconv.Atoi(errCode); err == nil {
				expectedExitCode = i
			}
		}

		// if it's exited, it's likely we'll need to clean it - we guard the "expensive" inspect call this way
		if cont.State == ExitedState {
			if !m.Quiet {
				log.Printf("Checking container cleanliness: %v (%v)\n", cont.ID, cont.Names[0])
			}

			inspect, err := m.Dockerd.Inspect(cont)
			if err != nil {
				log.Printf("Inspect failed: %v\n", err)
				continue
			}

			// if it's got the expected error code, we clean it up
			if inspect.State.ExitCode == expectedExitCode {
				if !m.Quiet {
					log.Printf("Found container to cleanup: %v (%v)\n", cont.ID, cont.Names[0])
				}
				if err := m.Dockerd.Remove(cont); err != nil {
					log.Printf("Failed to remove container %v (%v): %v\n", cont.ID, cont.Names[0], err)
				} else {
					log.Printf("Container cleaned: %v (%v)\n", cont.ID, cont.Names[0])
				}
			}
		}
	}
}

// Poll checks the dockerd system and executes operations as needed
func (m *Monitor) Poll(t time.Time) {
	if !m.Quiet {
		log.Printf("CheckStart for %v\n", t)
	}
	m.handleContainerHealth()
	m.handleContainerCleanup()
	if !m.Quiet {
		log.Printf("CheckEnd for %v\n", t)
	}
}
