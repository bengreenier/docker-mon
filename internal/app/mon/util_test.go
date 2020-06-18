package mon

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/testutil/assert"
)

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
