package worker

import (
	"Cube/task"
	"fmt"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Worker struct {
	Name      string
	Db        map[uuid.UUID]task.Task
	TaskCount int
	Queue     queue.Queue
}

func (w *Worker) RunTask() {
	fmt.Println("I will run task")
}

func (w *Worker) StartTask() {
	fmt.Println("I will Start Task")
}

func (w *Worker) StopTask() {
	fmt.Println("I will Start Task")
}

func (w *Worker) CollectStats() {
	fmt.Println("I will Start Task")
}
