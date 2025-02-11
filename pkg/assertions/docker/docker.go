package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// DockerAsserter implements assertions for Docker operations
type DockerAsserter struct {
	client DockerClient
}

func NewDockerAsserter(client DockerClient) *DockerAsserter {
	return &DockerAsserter{
		client: client,
	}
}

// DockerClient interface defines methods needed for Docker operations
type DockerClient interface {
	ImagePull(context.Context, string, types.ImagePullOptions) (io.ReadCloser, error)
	ContainerCreate(context.Context, *container.Config, *container.HostConfig, interface{}, interface{}, string) (container.ContainerCreateCreatedBody, error)
	ContainerStart(context.Context, string, types.ContainerStartOptions) error
	ContainerWait(context.Context, string, container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error)
	ContainerLogs(context.Context, string, types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerRemove(context.Context, string, types.ContainerRemoveOptions) error
}

// Runner handles Docker container operations
type Runner struct {
	client DockerClient
}

// NewRunner creates a new Docker runner instance
func NewRunner() (*Runner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}
	return &Runner{client: cli}, nil
}

// RunContainer executes a Docker container with the given image and command
func (r *Runner) RunContainer(ctx context.Context, image string, cmd []string) error {
	// Pull the image
	reader, err := r.client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)

	// Create container
	resp, err := r.client.ContainerCreate(ctx,
		&container.Config{
			Image: image,
			Cmd:   cmd,
		},
		nil, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := r.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to finish
	statusCh, errCh := r.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for container: %w", err)
		}
	case <-statusCh:
	}

	// Get container logs
	out, err := r.client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return fmt.Errorf("failed to get container logs: %w", err)
	}
	defer out.Close()

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	// Remove container
	err = r.client.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}
