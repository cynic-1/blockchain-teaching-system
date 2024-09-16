package docker

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerManager struct {
	client *client.Client
}

func NewDockerManager(apiVersion string) (*DockerManager, error) {
	cli, err := client.NewClientWithOpts(client.WithVersion(apiVersion))
	if err != nil {
		return nil, err
	}
	return &DockerManager{client: cli}, nil
}

func (dm *DockerManager) CreateContainer(ctx context.Context, image string, cmd []string) (string, error) {
	resp, err := dm.client.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   cmd,
	}, nil, nil, nil, "")
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (dm *DockerManager) StartContainer(ctx context.Context, containerID string) error {
	return dm.client.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (dm *DockerManager) StopContainer(ctx context.Context, containerID string) error {
	stopOptions := container.StopOptions{}
	return dm.client.ContainerStop(ctx, containerID, stopOptions)
}

func (dm *DockerManager) RemoveContainer(ctx context.Context, containerID string) error {
	removeOptions := container.RemoveOptions{
		Force: true,
	}
	return dm.client.ContainerRemove(ctx, containerID, removeOptions)
}
