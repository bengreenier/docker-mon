package mon

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/testutil/assert"
)

type mockContainerListResponse struct {
	err   error
	conts []types.Container
}

type mockDocker struct {
	listCallCount     int
	listCallParam0    []map[string]string
	listResponse      mockContainerListResponse
	restartCallCount  int
	restartCallParam0 []int64
	restartCallParam1 []types.Container
	restartResponse   error
	removeCallCount   int
	removeCallParam0  []types.Container
	removeResponse    error
}

const testPrefix string = "test_cont"

func TestHandleContainerHealthOk(t *testing.T) {
	dockerd := mockDocker{
		listResponse: mockContainerListResponse{
			err:   nil,
			conts: makeUnhealthyContainerList(),
		},
		restartResponse: nil,
		removeResponse:  nil,
	}

	mon := Monitor{
		ContainerPrefix: testPrefix,
		Dockerd:         &dockerd,
	}

	mon.handleContainerHealth()

	// should find and restart the unhealthy container
	// should only find containers with our testPrefix
	// should not remove anything
	assert.Equal(t, dockerd.listCallCount, 1)
	assert.DeepEqual(t, dockerd.listCallParam0[0], map[string]string{
		"label": fmt.Sprintf("%s,%s", ObserveLabel, CheckHealthLabel),
	})
	assert.Equal(t, dockerd.restartCallCount, 2)
	assert.Equal(t, dockerd.restartCallParam0[0], int64(1337))
	assert.Equal(t, dockerd.restartCallParam1[0].ID, "abc123")
	assert.Equal(t, dockerd.restartCallParam0[1], DefaultRestartTimeoutMs)
	assert.Equal(t, dockerd.restartCallParam1[1].ID, "ghi123")
	assert.Equal(t, dockerd.removeCallCount, 0)
}

func TestHandleContainerHealthListErr(t *testing.T) {
	dockerd := mockDocker{
		listResponse: mockContainerListResponse{
			err: errors.New("test failure"),
		},
		restartResponse: nil,
		removeResponse:  nil,
	}

	mon := Monitor{
		ContainerPrefix: testPrefix,
		Dockerd:         &dockerd,
	}

	mon.handleContainerHealth()

	assert.Equal(t, dockerd.listCallCount, 1)
	assert.Equal(t, dockerd.removeCallCount, 0)
	assert.Equal(t, dockerd.restartCallCount, 0)
}

func TestHandleContainerHealthErr(t *testing.T) {
	dockerd := mockDocker{
		listResponse: mockContainerListResponse{
			err:   nil,
			conts: makeUnhealthyContainerList(),
		},
		restartResponse: nil,
		removeResponse:  errors.New("test fail"),
	}

	mon := Monitor{
		ContainerPrefix: testPrefix,
		Dockerd:         &dockerd,
	}

	mon.handleContainerHealth()

	assert.Equal(t, dockerd.listCallCount, 1)
	assert.Equal(t, dockerd.removeCallCount, 0)
	assert.Equal(t, dockerd.restartCallCount, 2)
}

func TestHandleContainerCleanupOk(t *testing.T) {
	dockerd := mockDocker{
		listResponse: mockContainerListResponse{
			err:   nil,
			conts: makeDirtyContainerList(),
		},
		restartResponse: nil,
		removeResponse:  nil,
	}

	mon := Monitor{
		ContainerPrefix: testPrefix,
		Dockerd:         &dockerd,
	}

	mon.handleContainerCleanup()

	// should find and stop the dirty container
	// should only stop if the container is in the correct state
	// should only find containers with our testPrefix
	// should not restart anything
	assert.Equal(t, dockerd.listCallCount, 1)
	assert.DeepEqual(t, dockerd.listCallParam0[0], map[string]string{
		"label": fmt.Sprintf("%s,%s", ObserveLabel, CheckCleanupLabel),
	})
	assert.Equal(t, dockerd.restartCallCount, 0)
	assert.Equal(t, dockerd.removeCallCount, 2)
	assert.Equal(t, dockerd.removeCallParam0[0].ID, "abc123")
	assert.Equal(t, dockerd.removeCallParam0[1].ID, "ghi123")
}

func TestHandleContainerCleanupListErr(t *testing.T) {
	dockerd := mockDocker{
		listResponse: mockContainerListResponse{
			err: errors.New("test failure"),
		},
		restartResponse: nil,
		removeResponse:  nil,
	}

	mon := Monitor{
		ContainerPrefix: testPrefix,
		Dockerd:         &dockerd,
	}

	mon.handleContainerCleanup()

	assert.Equal(t, dockerd.listCallCount, 1)
	assert.Equal(t, dockerd.removeCallCount, 0)
	assert.Equal(t, dockerd.restartCallCount, 0)
}

func TestHandleContainerCleanupErr(t *testing.T) {
	dockerd := mockDocker{
		listResponse: mockContainerListResponse{
			err:   nil,
			conts: makeDirtyContainerList(),
		},
		restartResponse: nil,
		removeResponse:  errors.New("test fail"),
	}

	mon := Monitor{
		ContainerPrefix: testPrefix,
		Dockerd:         &dockerd,
	}

	mon.handleContainerCleanup()

	assert.Equal(t, dockerd.listCallCount, 1)
	assert.Equal(t, dockerd.removeCallCount, 2)
	assert.Equal(t, dockerd.restartCallCount, 0)
}

func TestPollOk(t *testing.T) {
	dockerd := mockDocker{
		listResponse: mockContainerListResponse{
			err:   errors.New("test fail"),
			conts: []types.Container{},
		},
		restartResponse: nil,
		removeResponse:  nil,
	}

	mon := Monitor{
		ContainerPrefix: testPrefix,
		Dockerd:         &dockerd,
	}

	mon.Poll(time.Now())

	// expect that list was called twice, once for each handler
	assert.Equal(t, dockerd.listCallCount, 2)
	assert.Equal(t, dockerd.removeCallCount, 0)
	assert.Equal(t, dockerd.restartCallCount, 0)
}

func (d *mockDocker) ExecuteListQuery(filterMap map[string]string) ([]types.Container, error) {
	d.listCallCount++
	d.listCallParam0 = append(d.listCallParam0, filterMap)
	return d.listResponse.conts, d.listResponse.err
}

func (d *mockDocker) Restart(timeoutMs int64, cont types.Container) error {
	d.restartCallCount++
	d.restartCallParam0 = append(d.restartCallParam0, timeoutMs)
	d.restartCallParam1 = append(d.restartCallParam1, cont)
	return d.restartResponse
}

func (d *mockDocker) Remove(cont types.Container) error {
	d.removeCallCount++
	d.removeCallParam0 = append(d.removeCallParam0, cont)
	return d.removeResponse
}

func makeUnhealthyContainerList() []types.Container {
	return []types.Container{
		{
			Labels: map[string]string{
				"mon.observe":               "1",
				"mon.checks.health":         "1",
				"mon.checks.health.timeout": "1337",
			},
			Names: []string{
				fmt.Sprintf("%vtest-container-1", testPrefix),
			},
			ID:    "abc123",
			State: makeUnhealthyContainerState(),
		},
		{
			Labels: map[string]string{
				"mon.observe":       "1",
				"mon.checks.health": "1",
			},
			Names: []string{
				"test-container-2",
			},
			ID:    "def123",
			State: makeUnhealthyContainerState(),
		},
		{
			Labels: map[string]string{
				"mon.observe":       "1",
				"mon.checks.health": "1",
			},
			Names: []string{
				fmt.Sprintf("%vtest-container-3", testPrefix),
			},
			ID:    "ghi123",
			State: makeUnhealthyContainerState(),
		},
		{
			Labels: map[string]string{
				"mon.observe":       "1",
				"mon.checks.health": "1",
			},
			Names: []string{
				fmt.Sprintf("%vtest-container-4", testPrefix),
			},
			ID:    "jkl123",
			State: makeHealthyContainerState(),
		},
	}
}

func makeDirtyContainerList() []types.Container {
	return []types.Container{
		{
			Labels: map[string]string{
				"mon.observe":        "1",
				"mon.checks.cleanup": "1",
			},
			Names: []string{
				fmt.Sprintf("%vtest-container-1", testPrefix),
			},
			ID:    "abc123",
			State: makeStoppedContainerState(0),
		},
		{
			Labels: map[string]string{
				"mon.observe":        "1",
				"mon.checks.cleanup": "1",
			},
			Names: []string{
				fmt.Sprintf("%vtest-container-2", testPrefix),
			},
			ID:    "def123",
			State: makeStoppedContainerState(2),
		},
		{
			Labels: map[string]string{
				"mon.observe":             "1",
				"mon.checks.cleanup":      "1",
				"mon.checks.cleanup.code": "3",
			},
			Names: []string{
				fmt.Sprintf("%vtest-container-3", testPrefix),
			},
			ID:    "ghi123",
			State: makeStoppedContainerState(3),
		},
	}
}

func makeUnhealthyContainerState() string {
	typed := types.ContainerState{
		Health: &types.Health{
			Status: Unhealthy,
		},
		Running: true,
	}

	dat, err := json.Marshal(typed)
	if err != nil {
		panic(err)
	}

	return string(dat)
}

func makeHealthyContainerState() string {
	typed := types.ContainerState{
		Health: &types.Health{
			Status: "Healthy",
		},
		Running: true,
	}

	dat, err := json.Marshal(typed)
	if err != nil {
		panic(err)
	}

	return string(dat)
}

func makeStoppedContainerState(ec int) string {
	typed := types.ContainerState{
		Running:    false,
		Restarting: false,
		ExitCode:   ec,
	}

	dat, err := json.Marshal(typed)
	if err != nil {
		panic(err)
	}

	return string(dat)
}
