package mon

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker/client"
	"github.com/docker/engine/api/types"
)

func parseState(state string) (types.ContainerState, error) {
	var dat types.ContainerState
	if err := json.Unmarshal([]byte(state), &dat); err != nil {
		return types.ContainerState{}, err
	}

	return dat, nil
}

func namesContainPrefix(names []string, prefix string) bool {
	for _, name := range names {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}

func withRetry(max int64, fn func() error) error {
	for i := int64(0); i < max; i++ {
		err := fn()

		if err == nil {
			return nil
		}

		fmt.Printf("Retrying (attempt %v/%v)...\n", i, max)
	}

	return errors.New("Retry count exceeded")
}

func withCli(controlAddr string, targetVer string, fn func(*client.Client) error) error {
	cli, err := client.NewClient(controlAddr, targetVer, nil, nil)
	if err != nil {
		return err
	}

	if err := fn(cli); err != nil {
		return err
	}

	if err := cli.Close(); err != nil {
		return err
	}

	return nil
}
