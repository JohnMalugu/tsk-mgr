package service

import (
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

// In-memory storage (we'll use database later)
var tasks []model.Task
var nextID int = 1

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
	return tasks
}

// GetTaskByID returns a task by ID
func GetTaskByID(id int) *model.Task {
	for i, task := range tasks {
		if task.ID == id {
			return &tasks[i]
		}
	}
	return nil
}

// CreateTask creates a new task
func CreateTask(task model.Task) model.Task {
	task.ID = nextID
	nextID++
	tasks = append(tasks, task)
	return task
}
