package service

import (
	"sync"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

// In-memory storage (we'll use database later)
var tasks []model.Task
var nextID int = 1
var mu sync.RWMutex

func init() {
	// Initialize sample data
	tasks = []model.Task{
		{ID: 1, Title: "Buy groceries", DueDate: time.Now().AddDate(0, 0, -1), Completed: false},
		{ID: 2, Title: "Learn Go", DueDate: time.Now().AddDate(0, 0, 3), Completed: false},
	}
	nextID = 3
}

// GetAllTasks returns all tasks
func GetAllTasks() []model.Task {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]model.Task, len(tasks))
	copy(result, tasks)
	return result
}

// GetTaskByID returns a task by ID
func GetTaskByID(id int) *model.Task {
	mu.RLock()
	defer mu.RUnlock()

	for i, task := range tasks {
		if task.ID == id {
			return &tasks[i]
		}
	}
	return nil
}

// CreateTask creates a new task
func CreateTask(task model.Task) model.Task {
	mu.Lock()
	defer mu.Unlock()

	task.ID = nextID
	nextID++
	tasks = append(tasks, task)
	return task
}

// UpdateTask replaces an existing task and returns the updated task.
func UpdateTask(id int, task model.Task) *model.Task {
	mu.Lock()
	defer mu.Unlock()

	for i := range tasks {
		if tasks[i].ID == id {
			task.ID = id
			tasks[i] = task
			updated := tasks[i]
			return &updated
		}
	}
	return nil
}

// DeleteTask removes an existing task.
func DeleteTask(id int) bool {
	mu.Lock()
	defer mu.Unlock()

	for i := range tasks {
		if tasks[i].ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			return true
		}
	}
	return false
}
