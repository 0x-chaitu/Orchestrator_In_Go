package task

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
)

type Task struct {
	ID            uuid.UUID
	State         State
	Name          string
	Image         string
	Memory        int
	Disk          int
	ExposedPorts  nat.PortSet
	PortBinding   map[string]string
	RestartPolicy string
	StartTime     time.Time
	FinishTime    time.Time
}

type TaskEvent struct {
	State     State
	ID        uuid.UUID
	TimeStamp time.Time
	Task      Task
}

type Config struct {
	Name          string
	AttachStdin   bool
	AttachStdout  bool
	AttachStdErr  bool
	Cmd           []string
	Image         string
	Memory        int64
	Disk          int64
	Env           []string
	RestartPolicy string
}

type Docker struct {
	Client      *client.Client
	Config      Config
	ContainerId string
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(ctx, d.Config.Image, types.ImagePullOptions{})
	if err != nil {
		log.Printf("Error pulling Image: %s %v", d.Config.Name, err)
		return DockerResult{Error: err}
	}
	io.Copy(os.Stdout, reader)
	rp := container.RestartPolicy{
		Name: d.Config.RestartPolicy,
	}
	r := container.Resources{
		Memory: d.Config.Memory,
	}

	cc := container.Config{
		Image: d.Config.Image,
		Env:   d.Config.Env,
	}

	hc := container.HostConfig{
		RestartPolicy:   rp,
		Resources:       r,
		PublishAllPorts: true,
	}

	resp, err := d.Client.ContainerCreate(ctx, &cc, &hc, nil, nil, d.Config.Name)
	if err != nil {
		log.Printf("Error while creating container: %s %v\n", d.Config.Image, err)
		return DockerResult{Error: err}
	}
	err = d.Client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Printf("Error in starting container: %s %v", d.Config.Image, err)
		return DockerResult{Error: err}
	}
	d.ContainerId = resp.ID
	out, err := d.Client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})

	if err != nil {
		log.Printf("Error getting logs for container %s %v\n", resp.ID, err)
		return DockerResult{
			Error: err,
		}
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return DockerResult{
		Result:      "success",
		Action:      "start",
		ContainerId: resp.ID,
	}
}

func (d *Docker) Stop() DockerResult {
	log.Printf("Attempting to stop container %s\n", d.ContainerId)
	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, d.ContainerId, container.StopOptions{})
	if err != nil {
		fmt.Printf("Error stopping container: %s %v", d.ContainerId, err)
		panic(err)
	}
	err = d.Client.ContainerRemove(ctx, d.ContainerId, types.ContainerRemoveOptions{})
	if err != nil {
		fmt.Printf("Error removing container %v\n", err)
		panic(err)
	}
	return DockerResult{ContainerId: d.ContainerId, Action: "stop", Result: "success"}

}
