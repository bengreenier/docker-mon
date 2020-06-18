package mon

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	mocks "github.com/bengreenier/docker-mon/internal/app/mon/mocks"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/golang/mock/gomock"
)

const testContainerNamePrefix string = "test_cont"

const nonstandardTimeoutValue int64 = 1337

// 4 cleanup cases, 4 health cases. incl 1 no-check case for both count.
var testData = []types.ContainerJSON{
	// exited, observable, cleanup with code = 0
	{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:   "abc",
			Name: fmt.Sprintf("%s_abc", testContainerNamePrefix),
			State: &types.ContainerState{
				Status:   ExitedState,
				ExitCode: 0,
			},
		},
		Config: &container.Config{
			Labels: map[string]string{
				"mon.observe":        "1",
				"mon.checks.cleanup": "1",
			},
		},
	},
	// exited, observable, cleanup with code = 1337
	{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:   "def",
			Name: fmt.Sprintf("%s_def", testContainerNamePrefix),
			State: &types.ContainerState{
				Status:   ExitedState,
				ExitCode: 1337,
			},
		},
		Config: &container.Config{
			Labels: map[string]string{
				"mon.observe":             "1",
				"mon.checks.cleanup":      "1",
				"mon.checks.cleanup.code": "1337",
			},
		},
	},
	// exited, observable, no checks
	{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:   "ghi",
			Name: fmt.Sprintf("%s_ghi", testContainerNamePrefix),
			State: &types.ContainerState{
				Status:   ExitedState,
				ExitCode: 0,
			},
		},
		Config: &container.Config{
			Labels: map[string]string{
				"mon.observe": "1",
			},
		},
	},
	// running, observable, cleanup with code = 0
	{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:   "jkl",
			Name: fmt.Sprintf("%s_jkl", testContainerNamePrefix),
			State: &types.ContainerState{
				Status: RunningState,
			},
		},
		Config: &container.Config{
			Labels: map[string]string{
				"mon.observe":        "1",
				"mon.checks.cleanup": "1",
			},
		},
	},
	// running, observable, healtcheck that is failing
	{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:   "mno",
			Name: fmt.Sprintf("%s_mno", testContainerNamePrefix),
			State: &types.ContainerState{
				Status: RunningState,
				Health: &types.Health{
					Status: types.Unhealthy,
				},
			},
		},
		Config: &container.Config{
			Labels: map[string]string{
				"mon.observe":       "1",
				"mon.checks.health": "1",
			},
		},
	},
	// running, observable, healtcheck that is failing with nonstandardTimeout
	{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:   "mno2",
			Name: fmt.Sprintf("%s_mno2", testContainerNamePrefix),
			State: &types.ContainerState{
				Status: RunningState,
				Health: &types.Health{
					Status: types.Unhealthy,
				},
			},
		},
		Config: &container.Config{
			Labels: map[string]string{
				"mon.observe":               "1",
				"mon.checks.health":         "1",
				"mon.checks.health.timeout": strconv.FormatInt(nonstandardTimeoutValue, 10),
			},
		},
	},
	// running, observable, healtcheck that is passing
	{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:   "pqr",
			Name: fmt.Sprintf("%s_pqr", testContainerNamePrefix),
			State: &types.ContainerState{
				Status: RunningState,
				Health: &types.Health{
					Status: types.Healthy,
				},
			},
		},
		Config: &container.Config{
			Labels: map[string]string{
				"mon.observe":       "1",
				"mon.checks.health": "1",
			},
		},
	},
	// running, observable, healtcheck configured, not set
	{
		ContainerJSONBase: &types.ContainerJSONBase{
			ID:   "stu",
			Name: fmt.Sprintf("%s_stu", testContainerNamePrefix),
			State: &types.ContainerState{
				Status: RunningState,
				Health: nil,
			},
		},
		Config: &container.Config{
			Labels: map[string]string{
				"mon.observe":       "1",
				"mon.checks.health": "1",
			},
		},
	},
}

// represent the testData as containers
var testContainers = jsonToContainers(testData)

func TestMonitorHandleCleanupOk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockDockerAPI(ctrl)

	m.
		EXPECT().
		ExecuteListQuery(gomock.Eq([]string{
			ObserveLabel,
			CheckCleanupLabel,
		})).
		Return(filterContainers(map[string]string{
			"mon.observe":        "1",
			"mon.checks.cleanup": "1",
		}, testContainers), nil)
	m.
		EXPECT().
		Inspect(gomock.Eq(testContainers[0])).
		Times(1).
		Return(testData[0], nil)
	m.
		EXPECT().
		Inspect(gomock.Eq(testContainers[1])).
		Times(1).
		Return(testData[1], nil)
	m.
		EXPECT().
		Remove(gomock.Eq(testContainers[0])).
		Times(1).
		Return(nil)
	m.
		EXPECT().
		Remove(gomock.Eq(testContainers[1])).
		Times(1).
		Return(nil)

	monitor := Monitor{
		ContainerPrefix: testContainerNamePrefix,
		Dockerd:         m,
	}

	monitor.handleContainerCleanup()
}

func TestMonitorHandleHealthCheckOk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockDockerAPI(ctrl)

	m.
		EXPECT().
		ExecuteListQuery(gomock.Eq([]string{
			ObserveLabel,
			CheckHealthLabel,
		})).
		Return(filterContainers(map[string]string{
			"mon.observe":       "1",
			"mon.checks.health": "1",
		}, testContainers), nil)
	m.
		EXPECT().
		Inspect(gomock.Eq(testContainers[4])).
		Times(1).
		Return(testData[4], nil)
	m.
		EXPECT().
		Inspect(gomock.Eq(testContainers[5])).
		Times(1).
		Return(testData[5], nil)
	m.
		EXPECT().
		Inspect(gomock.Eq(testContainers[6])).
		Times(1).
		Return(testData[6], nil)
	m.
		EXPECT().
		Inspect(gomock.Eq(testContainers[7])).
		Times(1).
		Return(testData[7], nil)
	m.
		EXPECT().
		Restart(gomock.Eq(DefaultRestartTimeoutMs), gomock.Eq(testContainers[4])).
		Times(1).
		Return(nil)
	m.
		EXPECT().
		Restart(gomock.Eq(nonstandardTimeoutValue), gomock.Eq(testContainers[5])).
		Times(1).
		Return(nil)

	monitor := Monitor{
		ContainerPrefix: testContainerNamePrefix,
		Dockerd:         m,
	}

	monitor.handleContainerHealth()
}

func TestMonitorPollOk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockDockerAPI(ctrl)

	m.
		EXPECT().
		ExecuteListQuery(gomock.Eq([]string{
			ObserveLabel,
			CheckCleanupLabel,
		})).
		Times(1).
		Return([]types.Container{}, errors.New("test failure"))
	m.
		EXPECT().
		ExecuteListQuery(gomock.Eq([]string{
			ObserveLabel,
			CheckHealthLabel,
		})).
		Times(1).
		Return([]types.Container{}, errors.New("test failure"))

	monitor := Monitor{
		ContainerPrefix: testContainerNamePrefix,
		Dockerd:         m,
	}

	// will swallow the errors
	monitor.Poll(time.Now())
}

// jsonToSingleContainer does a non-production-quality mapping between the types
func jsonToSingleContainer(json types.ContainerJSON) types.Container {
	return types.Container{
		ID:     json.ID,
		Names:  []string{json.Name},
		Labels: json.Config.Labels,
		State:  json.State.Status,
		Status: json.State.Status,
	}
}

// jsonToContainer does a non-production-quality mapping between the types as arrays
func jsonToContainers(json []types.ContainerJSON) []types.Container {
	res := make([]types.Container, len(json))
	for i, val := range json {
		res[i] = jsonToSingleContainer(val)
	}
	return res
}

// filterContainers limits a list to elements containing all given filters
func filterContainers(filterMap map[string]string, conts []types.Container) []types.Container {
	var res []types.Container

	for _, cont := range conts {
		matchesAll := true
		for k, v := range filterMap {
			if iv, ok := cont.Labels[k]; !ok || v != iv {
				matchesAll = false
			}
		}
		if matchesAll {
			res = append(res, cont)
		}
	}

	return res
}
