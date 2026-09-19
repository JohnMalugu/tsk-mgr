package service

import (
	"strings"
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

// GetAllTasks returns the first page of all tasks.
func GetAllTasks() []model.Task {
	return GetTasks(nil, "", 0, 20)
}

// GetTasks returns a page of tasks matching the optional filters.
func GetTasks(completed *bool, search string, offset, limit int) []model.Task {
	mu.RLock()
	defer mu.RUnlock()

	search = strings.ToLower(strings.TrimSpace(search))
	result := make([]model.Task, 0, len(tasks))
	for _, task := range tasks {
		if completed != nil && task.Completed != *completed {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(task.Title), search) {
			continue
		}
		result = append(result, task)
	}
	if offset >= len(result) {
		return []model.Task{}
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end]
}

// GetTaskByID returns a task by ID
func GetTaskByID(id int) *model.Task {
	mu.RLock()
	defer mu.RUnlock()

	for i, task := range tasks {
		if task.ID == id {
			result := tasks[i]
			return &result
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
