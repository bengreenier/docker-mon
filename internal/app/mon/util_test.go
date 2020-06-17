package mon

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/testutil/assert"
)

func TestParseStateOk(t *testing.T) {
	input := types.ContainerState{
		Health: &types.Health{
			Status: "Healthy",
		},
		Running:  true,
		ExitCode: 0,
	}

	output, err := parseState(serializeState(input))

	assert.Equal(t, err, nil)
	assert.Equal(t, output.Health.Status, input.Health.Status)
	assert.Equal(t, output.Health.FailingStreak, input.Health.FailingStreak)

	assert.Equal(t, output.Dead, input.Dead)
	assert.Equal(t, output.Error, input.Error)
	assert.Equal(t, output.ExitCode, input.ExitCode)
	assert.Equal(t, output.FinishedAt, input.FinishedAt)
	assert.Equal(t, output.OOMKilled, input.OOMKilled)
	assert.Equal(t, output.Paused, input.Paused)
	assert.Equal(t, output.Pid, input.Pid)
	assert.Equal(t, output.Restarting, input.Restarting)
	assert.Equal(t, output.Running, input.Running)
	assert.Equal(t, output.StartedAt, input.StartedAt)
	assert.Equal(t, output.Status, input.Status)
}

func TestParseStateErr(t *testing.T) {
	_, err := parseState("invalid")

	assert.NotNil(t, err)
}

func TestNamesContainPrefix(t *testing.T) {
	assert.Equal(t, namesContainPrefix([]string{"a_one"}, "a"), true)
	assert.Equal(t, namesContainPrefix([]string{"a_one", "a_two"}, "a"), true)
	assert.Equal(t, namesContainPrefix([]string{"a_one", "two"}, "a"), true)
	assert.Equal(t, namesContainPrefix([]string{"a_one", "a_two"}, "b"), false)
	assert.Equal(t, namesContainPrefix([]string{"a_one"}, "b"), false)
}

func TestWithRetryOk(t *testing.T) {
	count := 0
	err := withRetry(5, func() error {
		count++
		return nil
	})

	assert.NilError(t, err)
	assert.Equal(t, count, 1)
}

func TestWithRetryErr(t *testing.T) {
	count := 0
	err := withRetry(5, func() error {
		count++
		return errors.New("test fail")
	})

	assert.NotNil(t, err)
	assert.Equal(t, count, 5)
}

func serializeState(state types.ContainerState) string {
	dat, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}

	return string(dat)
}
