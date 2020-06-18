package mon

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// DockerAPI is something that implements the docker API
type DockerAPI interface {
	ExecuteListQuery(filterList []string) ([]types.Container, error)
	Restart(timeoutMs int64, cont types.Container) error
	Remove(types.Container) error
	Inspect(cont types.Container) (types.ContainerJSON, error)
}

// DockerD implements the DockerAPI for the docker daemon
type DockerD struct {
	ControlAddr    string
	TargetVersion  string
	CommandRetries int64
}

func (d *DockerD) withRetry(fn func() error) error {
	// we wrap this so the caller doesn't need to pass d.CommandRetries around
	return withRetry(d.CommandRetries, fn)
}

func (d *DockerD) withCli(fn func(*client.Client) error) error {
	// we wrap this so the caller doesn't need to pass addr, ver around
	return withCli(d.ControlAddr, d.TargetVersion, fn)
}

// ExecuteListQuery to find containers
func (d *DockerD) ExecuteListQuery(filterList []string) ([]types.Container, error) {
	filterArgs := filters.NewArgs()

	for _, val := range filterList {
		filterArgs.Add("label", val)
	}

	var containerList []types.Container

	ctx := context.Background()

	if err := d.withRetry(func() error {
		return d.withCli(func(cli *client.Client) error {
			data, err := cli.ContainerList(ctx, types.ContainerListOptions{
				All:     true,
				Filters: filterArgs,
			})

			if err != nil {
				return err
			}

			containerList = data

			return nil
		})
	}); err != nil {
		return nil, err
	}

	return containerList, nil
}

// Restart a container
func (d *DockerD) Restart(timeoutMs int64, cont types.Container) error {
	ctx := context.Background()

	return d.withRetry(func() error {
		return d.withCli(func(cli *client.Client) error {
			duration := time.Duration(timeoutMs) * time.Millisecond
			return cli.ContainerRestart(ctx, cont.ID, &duration)
		})
	})
}

// Remove a container
func (d *DockerD) Remove(cont types.Container) error {
	ctx := context.Background()

	return d.withRetry(func() error {
		return d.withCli(func(cli *client.Client) error {
			return cli.ContainerRemove(ctx, cont.ID, types.ContainerRemoveOptions{})
		})
	})
}

// Inspect a container
func (d *DockerD) Inspect(cont types.Container) (types.ContainerJSON, error) {
	ctx := context.Background()

	var data types.ContainerJSON

	if err := d.withRetry(func() error {
		return d.withCli(func(cli *client.Client) error {
			output, err := cli.ContainerInspect(ctx, cont.ID)
			if err != nil {
				return err
			}

			data = output

			return nil
		})
	}); err != nil {
		return types.ContainerJSON{}, err
	}

	return data, nil
}
