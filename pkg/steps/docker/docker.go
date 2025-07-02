package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Docker represents a Docker client and associated configuration
type Docker struct {
	client  *client.Client
	workdir string
}

// New creates a new Docker client instance
func New(workdir string) (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Docker{
		client:  cli,
		workdir: workdir,
	}, nil
}

// Run executes a Docker container with the specified configuration
func (d *Docker) Run(ctx context.Context, image string, cmd []string, env []string, volumes map[string]string) error {
	// Pull the image if it doesn't exist locally
	reader, err := d.client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", image, err)
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)

	// Prepare volume bindings
	var binds []string
	for host, container := range volumes {
		hostPath, err := filepath.Abs(host)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", host, err)
		}
		binds = append(binds, fmt.Sprintf("%s:%s", hostPath, container))
	}

	// Create container
	resp, err := d.client.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   cmd,
		Env:   env,
		Tty:   false,
	}, &container.HostConfig{
		Binds: binds,
	}, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to finish and capture output
	statusCh, errCh := d.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for container: %w", err)
		}
	case <-statusCh:
	}

	// Get container logs
	out, err := d.client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return fmt.Errorf("failed to get container logs: %w", err)
	}
	defer out.Close()

	// Copy logs to stdout/stderr
	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	if err != nil {
		return fmt.Errorf("failed to copy container output: %w", err)
	}

	return nil
}

// Close closes the Docker client connection
func (d *Docker) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}
