package main

import (
	"Cube/manager"
	"Cube/node"
	"Cube/task"
	"Cube/worker"
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

func main() {
	t := task.Task{
		ID:     uuid.New(),
		State:  task.Pending,
		Image:  "Image - 1",
		Disk:   1,
		Memory: 1024,
	}

	te := task.TaskEvent{
		ID:        uuid.New(),
		State:     task.Pending,
		TimeStamp: time.Now(),
		Task:      t,
	}

	fmt.Printf("task: %v\n", t)
	fmt.Printf("task event: %v\n", te)

	w := worker.Worker{
		Queue: *queue.New(),
		Db:    make(map[uuid.UUID]task.Task),
	}

	fmt.Printf("worker: %v\n", w)
	w.CollectStats()
	w.RunTask()
	w.StartTask()
	w.StopTask()

	m := manager.Manager{
		Pending: *queue.New(),
		TaskDb:  make(map[string][]task.Task),
		EventDB: make(map[string][]task.TaskEvent),
		Workers: []string{w.Name},
	}

	fmt.Printf("maneger: %v\n", m)
	m.SelectWorker()
	m.UpdateTasks()
	m.SendWork()

	n := node.Node{
		Name:   "Node-1",
		Ip:     "127.0.0.1",
		Cores:  4,
		Memory: 1024,
		Disk:   25,
		Role:   "Worker",
	}

	fmt.Printf("node : %v\n", n)

	dockerTask, createResult := createContainer()
	if createResult.Error != nil {
		fmt.Println(createResult.Error)
		os.Exit(1)
	}
	time.Sleep(time.Second * 5)
	fmt.Printf("Stopping container %s: \n", Stop(dockerTask))
}

func createContainer() (*task.Docker, *task.DockerResult) {
	c := task.Config{
		Name:  "postgres-1",
		Image: "postgres:13",
		Env: []string{
			"POSTGRES_USER=cube",
			"POSTGRES_PASSWORD=cube",
		},
	}
	dc, _ := client.NewClientWithOpts(client.FromEnv)
	d := task.Docker{
		Config: c,
		Client: dc,
	}

	result := d.Run()
	if result.Error != nil {
		fmt.Printf("Error in creating container: %v", result.Error)
		return nil, nil
	}
	fmt.Printf("Container %s is running with config %v\n", c.Name, d)
	return &d, &result
}

func Stop(d *task.Docker) *task.DockerResult {
	result := d.Stop()
	if result.Error != nil {
		fmt.Printf("Error in stopping container: %v", result.Error)
		return nil
	}
	fmt.Printf("Stopped container %s:", d.ContainerId)
	return &result

}
