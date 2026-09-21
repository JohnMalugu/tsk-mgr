package service

import (
	"sort"
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
	return GetTasks(nil, "", 0, 20, "id", false)
}

// GetTasks returns a page of tasks matching the optional filters.
func GetTasks(completed *bool, search string, offset, limit int, sortBy string, descending bool) []model.Task {
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
	sort.SliceStable(result, func(i, j int) bool {
		comparison := 0
		switch sortBy {
		case "title":
			left, right := strings.ToLower(result[i].Title), strings.ToLower(result[j].Title)
			if left < right {
				comparison = -1
			} else if left > right {
				comparison = 1
			}
		case "dueDate":
			if result[i].DueDate.Before(result[j].DueDate) {
				comparison = -1
			} else if result[i].DueDate.After(result[j].DueDate) {
				comparison = 1
			}
		default:
			if result[i].ID < result[j].ID {
				comparison = -1
			} else if result[i].ID > result[j].ID {
				comparison = 1
			}
		}
		if comparison == 0 {
			comparison = result[i].ID - result[j].ID
		}
		if descending {
			return comparison > 0
		}
		return comparison < 0
	})
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

// CompleteTask marks a task as completed and returns it.
func CompleteTask(id int) *model.Task {
	return SetTaskCompletion(id, true)
}

// SetTaskCompletion updates the completion state of a task and returns it.
func SetTaskCompletion(id int, completed bool) *model.Task {
	mu.Lock()
	defer mu.Unlock()

	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Completed = completed
			updated := tasks[i]
			return &updated
		}
	}
	return nil
}
