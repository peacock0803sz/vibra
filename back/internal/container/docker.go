package container

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/peacock0803sz/vibra/back/internal/adapter"
)

// DockerRuntime implements Runtime using the Docker Engine API.
type DockerRuntime struct {
	cli *client.Client
}

// NewDockerRuntime creates a DockerRuntime from environment settings.
func NewDockerRuntime() (*DockerRuntime, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return &DockerRuntime{cli: cli}, nil
}

func (d *DockerRuntime) Ping(ctx context.Context) error {
	_, err := d.cli.Ping(ctx)
	return err
}

func (d *DockerRuntime) Create(ctx context.Context, spec *adapter.ContainerSpec) (string, error) {
	var mounts []mount.Mount
	for _, v := range spec.Volumes {
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   v.HostPath,
			Target:   v.ContainerPath,
			ReadOnly: v.ReadOnly,
		})
	}

	resp, err := d.cli.ContainerCreate(ctx,
		&container.Config{
			Image:      spec.Image,
			Cmd:        spec.Command,
			Env:        spec.Env,
			WorkingDir: spec.WorkDir,
			Tty:        false,
		},
		&container.HostConfig{
			Mounts:      mounts,
			NetworkMode: container.NetworkMode(spec.NetworkMode),
			Resources: container.Resources{
				CPUQuota: spec.CPUQuota,
				Memory:   spec.MemoryLimit,
			},
		},
		nil, nil, "",
	)
	if err != nil {
		return "", fmt.Errorf("container create: %w", err)
	}
	return resp.ID, nil
}

func (d *DockerRuntime) Start(ctx context.Context, id string) error {
	return d.cli.ContainerStart(ctx, id, container.StartOptions{})
}

// Logs decodes Docker's multiplexed stream and returns stdout only.
func (d *DockerRuntime) Logs(ctx context.Context, id string) (io.ReadCloser, error) {
	raw, err := d.cli.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return nil, err
	}

	// Demultiplex Docker's stream header format into clean stdout.
	pr, pw := io.Pipe()
	go func() {
		_, err := stdcopy.StdCopy(pw, io.Discard, raw)
		raw.Close()
		pw.CloseWithError(err)
	}()

	return pr, nil
}

func (d *DockerRuntime) Kill(ctx context.Context, id string) error {
	return d.cli.ContainerKill(ctx, id, "KILL")
}

func (d *DockerRuntime) Remove(ctx context.Context, id string) error {
	return d.cli.ContainerRemove(ctx, id, container.RemoveOptions{
		Force: true,
	})
}
