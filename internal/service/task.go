package service

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrTaskNotFound            = errors.New("task not found")
	ErrDependencySelfReference = errors.New("task cannot depend on itself")
	ErrDependencyAlreadyExists = errors.New("dependency already exists")
	ErrDependencyNotFound      = errors.New("dependency not found")
	ErrDependencyCycle         = errors.New("dependency would create a cycle")
	ErrTaskIsPrerequisite      = errors.New("task is a prerequisite for other tasks")
	ErrTaskBlocked             = errors.New("task has incomplete prerequisites")
)

// In-memory storage (we'll use database later)
var tasks []model.Task
var nextID int = 1
var nextChecklistItemID int = 1
var mu sync.RWMutex

func init() {
	ResetTasks()
}

// ResetTasks restores the default in-memory task seed.
func ResetTasks() {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	tasks = []model.Task{
		{ID: 1, Title: "Buy groceries", CreatedAt: now, UpdatedAt: now, DueDate: now.AddDate(0, 0, -1), Completed: false, Priority: "medium", Tags: []string{"home", "errands"}},
		{ID: 2, Title: "Learn Go", CreatedAt: now, UpdatedAt: now, DueDate: now.AddDate(0, 0, 3), Completed: false, Priority: "low", Tags: []string{"study"}},
	}
	nextID = 3
	nextChecklistItemID = 1
}

type TaskSummary struct {
	Total      int            `json:"total"`
	Completed  int            `json:"completed"`
	Pending    int            `json:"pending"`
	Overdue    int            `json:"overdue"`
	Blocked    int            `json:"blocked"`
	ByPriority map[string]int `json:"byPriority"`
}

type BulkUpdateResult struct {
	Tasks   []model.Task `json:"tasks"`
	Updated int          `json:"updated"`
}

type BulkDeleteResult struct {
	Deleted []int `json:"deleted"`
	Count   int   `json:"count"`
}

// GetAllTasks returns the first page of all tasks.
func GetAllTasks() []model.Task {
	return GetTasks(nil, nil, "", nil, nil, nil, nil, 0, 20, "id", false)
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func normalizeTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		key := strings.ToLower(tag)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, tag)
	}
	return result
}

// GetTaskSummary returns aggregate task counts.
func GetTaskSummary() TaskSummary {
	mu.RLock()
	defer mu.RUnlock()

	summary := TaskSummary{Total: len(tasks), ByPriority: map[string]int{"low": 0, "medium": 0, "high": 0}}
	for _, task := range tasks {
		if hasIncompletePrerequisiteLocked(task) {
			summary.Blocked++
		}
		priority := strings.ToLower(task.Priority)
		if _, ok := summary.ByPriority[priority]; ok {
			summary.ByPriority[priority]++
		}
		if task.Completed {
			summary.Completed++
			continue
		}
		summary.Pending++
		if task.DueDate.Before(time.Now()) {
			summary.Overdue++
		}
	}
	return summary
}

// GetTasks returns a page of tasks matching the optional filters.
func GetTasks(completed, overdue *bool, search string, priority *string, tag *string, dueAfter, dueBefore *time.Time, offset, limit int, sortBy string, descending bool) []model.Task {
	mu.RLock()
	defer mu.RUnlock()

	search = strings.ToLower(strings.TrimSpace(search))
	result := make([]model.Task, 0, len(tasks))
	for _, task := range tasks {
		if completed != nil && task.Completed != *completed {
			continue
		}
		isOverdue := !task.Completed && task.DueDate.Before(time.Now())
		if overdue != nil && isOverdue != *overdue {
			continue
		}
		if priority != nil && strings.ToLower(task.Priority) != strings.ToLower(*priority) {
			continue
		}
		if tag != nil && !contains(task.Tags, *tag) {
			continue
		}
		if dueAfter != nil && task.DueDate.Before(*dueAfter) {
			continue
		}
		if dueBefore != nil && task.DueDate.After(*dueBefore) {
			continue
		}
		matchesSearch := strings.Contains(strings.ToLower(task.Title), search)
		if search != "" && !matchesSearch {
			for _, tag := range task.Tags {
				if strings.Contains(strings.ToLower(tag), search) {
					matchesSearch = true
					break
				}
			}
		}
		if search != "" && !matchesSearch {
			continue
		}
		result = append(result, task)
	}
	sort.SliceStable(result, func(i, j int) bool {
		comparison := 0
		switch sortBy {
		case "completed":
			if !result[i].Completed && result[j].Completed {
				comparison = -1
			} else if result[i].Completed && !result[j].Completed {
				comparison = 1
			}
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
		case "priority":
			priorityRank := func(priority string) int {
				switch strings.ToLower(priority) {
				case "low":
					return 1
				case "medium":
					return 2
				case "high":
					return 3
				default:
					return 0
				}
			}
			left, right := priorityRank(result[i].Priority), priorityRank(result[j].Priority)
			if left < right {
				comparison = -1
			} else if left > right {
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

// GetTaskDependencies returns copies of all prerequisite tasks for a task ID.
func GetTaskDependencies(id int) ([]model.Task, bool) {
	mu.RLock()
	defer mu.RUnlock()

	var task *model.Task
	for i := range tasks {
		if tasks[i].ID == id {
			task = &tasks[i]
			break
		}
	}
	if task == nil {
		return nil, false
	}

	dependencies := make([]model.Task, 0, len(task.DependsOn))
	for _, dependencyID := range task.DependsOn {
		for _, candidate := range tasks {
			if candidate.ID == dependencyID {
				dependencies = append(dependencies, candidate)
				break
			}
		}
	}
	return dependencies, true
}

// IsTaskBlocked reports whether a task has any incomplete prerequisite.
func IsTaskBlocked(id int) (bool, bool) {
	mu.RLock()
	defer mu.RUnlock()

	for _, task := range tasks {
		if task.ID != id {
			continue
		}
		for _, dependencyID := range task.DependsOn {
			for _, dependency := range tasks {
				if dependency.ID == dependencyID && !dependency.Completed {
					return true, true
				}
			}
		}
		return false, true
	}
	return false, false
}

// GetReadyTasks returns incomplete tasks whose prerequisites are all complete.
func GetReadyTasks() []model.Task {
	mu.RLock()
	defer mu.RUnlock()

	ready := make([]model.Task, 0)
	for _, task := range tasks {
		if task.Completed || hasIncompletePrerequisiteLocked(task) {
			continue
		}
		copy := task
		copy.DependsOn = append([]int(nil), task.DependsOn...)
		ready = append(ready, copy)
	}
	sort.Slice(ready, func(i, j int) bool {
		return ready[i].ID < ready[j].ID
	})
	return ready
}

// AddTaskDependency makes dependencyID a prerequisite of taskID.
func AddTaskDependency(taskID, dependencyID int) (*model.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	taskPosition, dependencyPosition := -1, -1
	for i := range tasks {
		if tasks[i].ID == taskID {
			taskPosition = i
		}
		if tasks[i].ID == dependencyID {
			dependencyPosition = i
		}
	}
	if taskPosition < 0 || dependencyPosition < 0 {
		return nil, ErrTaskNotFound
	}
	if taskID == dependencyID {
		return nil, ErrDependencySelfReference
	}
	if tasks[taskPosition].Completed && !tasks[dependencyPosition].Completed {
		return nil, ErrTaskBlocked
	}
	for _, existing := range tasks[taskPosition].DependsOn {
		if existing == dependencyID {
			return nil, ErrDependencyAlreadyExists
		}
	}

	visited := make(map[int]struct{})
	var reachesTask func(int) bool
	reachesTask = func(currentID int) bool {
		if currentID == taskID {
			return true
		}
		if _, seen := visited[currentID]; seen {
			return false
		}
		visited[currentID] = struct{}{}
		for _, current := range tasks {
			if current.ID == currentID {
				for _, prerequisiteID := range current.DependsOn {
					if reachesTask(prerequisiteID) {
						return true
					}
				}
				break
			}
		}
		return false
	}
	if reachesTask(dependencyID) {
		return nil, ErrDependencyCycle
	}

	tasks[taskPosition].DependsOn = append(tasks[taskPosition].DependsOn, dependencyID)
	tasks[taskPosition].UpdatedAt = time.Now()
	updated := tasks[taskPosition]
	updated.DependsOn = append([]int(nil), updated.DependsOn...)
	return &updated, nil
}

// RemoveTaskDependency removes dependencyID from taskID's prerequisites.
func RemoveTaskDependency(taskID, dependencyID int) (*model.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	for i := range tasks {
		if tasks[i].ID != taskID {
			continue
		}
		for dependencyIndex, existingID := range tasks[i].DependsOn {
			if existingID != dependencyID {
				continue
			}
			tasks[i].DependsOn = append(tasks[i].DependsOn[:dependencyIndex], tasks[i].DependsOn[dependencyIndex+1:]...)
			tasks[i].UpdatedAt = time.Now()
			updated := tasks[i]
			updated.DependsOn = append([]int(nil), updated.DependsOn...)
			return &updated, nil
		}
		return nil, ErrDependencyNotFound
	}
	return nil, ErrTaskNotFound
}

// ReplaceTaskDependencies replaces a task's prerequisite set after validating graph integrity.
func ReplaceTaskDependencies(taskID int, dependencyIDs []int) (*model.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	taskPosition := -1
	positions := make(map[int]int, len(tasks))
	for i := range tasks {
		positions[tasks[i].ID] = i
		if tasks[i].ID == taskID {
			taskPosition = i
		}
	}
	if taskPosition < 0 {
		return nil, ErrTaskNotFound
	}
	validated := make([]int, 0, len(dependencyIDs))
	seen := make(map[int]struct{}, len(dependencyIDs))
	for _, dependencyID := range dependencyIDs {
		if dependencyID == taskID {
			return nil, ErrDependencySelfReference
		}
		if _, duplicate := seen[dependencyID]; duplicate {
			return nil, ErrDependencyAlreadyExists
		}
		if _, exists := positions[dependencyID]; !exists {
			return nil, ErrTaskNotFound
		}
		seen[dependencyID] = struct{}{}
		validated = append(validated, dependencyID)
	}
	if tasks[taskPosition].Completed && hasIncompleteDependencyIDsLocked(validated) {
		return nil, ErrTaskBlocked
	}

	original := tasks[taskPosition].DependsOn
	tasks[taskPosition].DependsOn = validated
	for _, dependencyID := range validated {
		visited := make(map[int]struct{})
		var reachesTask func(int) bool
		reachesTask = func(currentID int) bool {
			if currentID == taskID {
				return true
			}
			if _, visitedAlready := visited[currentID]; visitedAlready {
				return false
			}
			visited[currentID] = struct{}{}
			position, exists := positions[currentID]
			if !exists {
				return false
			}
			for _, prerequisiteID := range tasks[position].DependsOn {
				if reachesTask(prerequisiteID) {
					return true
				}
			}
			return false
		}
		if reachesTask(dependencyID) {
			tasks[taskPosition].DependsOn = original
			return nil, ErrDependencyCycle
		}
	}
	tasks[taskPosition].UpdatedAt = time.Now()
	updated := tasks[taskPosition]
	updated.DependsOn = append([]int(nil), updated.DependsOn...)
	return &updated, nil
}

// CreateTask creates a new task
func CreateTask(task model.Task) model.Task {
	created, _ := CreateTaskWithDependencies(task, task.DependsOn)
	return created
}

// CreateTaskWithDependencies creates a task and validates its prerequisites atomically.
func CreateTaskWithDependencies(task model.Task, dependencyIDs []int) (model.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	for _, dependencyID := range dependencyIDs {
		if dependencyID < 1 || findTaskPositionLocked(dependencyID) < 0 {
			return model.Task{}, ErrTaskNotFound
		}
	}
	seen := make(map[int]struct{}, len(dependencyIDs))
	for _, dependencyID := range dependencyIDs {
		if _, duplicate := seen[dependencyID]; duplicate {
			return model.Task{}, ErrDependencyAlreadyExists
		}
		seen[dependencyID] = struct{}{}
	}
	if task.Completed && hasIncompleteDependencyIDsLocked(dependencyIDs) {
		return model.Task{}, ErrTaskBlocked
	}
	now := time.Now()
	task.Tags = normalizeTags(task.Tags)
	if task.Priority == "" {
		task.Priority = "medium"
	}
	task.CreatedAt = now
	task.UpdatedAt = now
	task.ID = nextID
	task.DependsOn = append([]int(nil), dependencyIDs...)
	nextID++
	tasks = append(tasks, task)
	return task, nil
}

// UpdateTask replaces an existing task and returns the updated task.
func UpdateTask(id int, task model.Task) *model.Task {
	updated, _ := UpdateTaskWithDependencies(id, task, task.DependsOn)
	return updated
}

// UpdateTaskWithDependencies atomically replaces task fields and validates prerequisites.
func UpdateTaskWithDependencies(id int, task model.Task, dependencyIDs []int) (*model.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	for i := range tasks {
		if tasks[i].ID == id {
			if err := validateDependencySetLocked(id, dependencyIDs); err != nil {
				return nil, err
			}
			if task.Completed && hasIncompleteDependencyIDsLocked(dependencyIDs) {
				return nil, ErrTaskBlocked
			}
			if !task.Completed && hasCompletedDependentLocked(id, nil) {
				return nil, ErrTaskBlocked
			}
			task.CreatedAt = tasks[i].CreatedAt
			task.UpdatedAt = time.Now()
			task.Tags = normalizeTags(task.Tags)
			if task.Priority == "" {
				task.Priority = "medium"
			}
			task.ID = id
			task.DependsOn = append([]int(nil), dependencyIDs...)
			tasks[i] = task
			updated := tasks[i]
			return &updated, nil
		}
	}
	return nil, ErrTaskNotFound
}

func hasCompletedDependentLocked(taskID int, selected map[int]struct{}) bool {
	for _, task := range tasks {
		if !task.Completed {
			continue
		}
		if _, willBeUncompleted := selected[task.ID]; willBeUncompleted {
			continue
		}
		for _, dependencyID := range task.DependsOn {
			if dependencyID == taskID {
				return true
			}
		}
	}
	return false
}
func findTaskPositionLocked(id int) int {
	for i := range tasks {
		if tasks[i].ID == id {
			return i
		}
	}
	return -1
}

func validateDependencySetLocked(taskID int, dependencyIDs []int) error {
	if findTaskPositionLocked(taskID) < 0 {
		return ErrTaskNotFound
	}
	seen := make(map[int]struct{}, len(dependencyIDs))
	for _, dependencyID := range dependencyIDs {
		if dependencyID == taskID {
			return ErrDependencySelfReference
		}
		if _, duplicate := seen[dependencyID]; duplicate {
			return ErrDependencyAlreadyExists
		}
		if findTaskPositionLocked(dependencyID) < 0 {
			return ErrTaskNotFound
		}
		seen[dependencyID] = struct{}{}
	}
	visited := make(map[int]struct{})
	var reachesTask func(int) bool
	reachesTask = func(currentID int) bool {
		if currentID == taskID {
			return true
		}
		if _, seen := visited[currentID]; seen {
			return false
		}
		visited[currentID] = struct{}{}
		position := findTaskPositionLocked(currentID)
		if position < 0 {
			return false
		}
		for _, prerequisiteID := range tasks[position].DependsOn {
			if reachesTask(prerequisiteID) {
				return true
			}
		}
		return false
	}
	for _, dependencyID := range dependencyIDs {
		if reachesTask(dependencyID) {
			return ErrDependencyCycle
		}
	}
	return nil
}

// DeleteTask removes an existing task.
func DeleteTask(id int) bool {
	mu.Lock()
	defer mu.Unlock()

	for _, task := range tasks {
		for _, dependencyID := range task.DependsOn {
			if dependencyID == id {
				return false
			}
		}
	}
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
	updated, _ := SetTaskCompletionChecked(id, completed)
	return updated
}

// SetTaskCompletionChecked atomically updates completion and validates prerequisites.
func SetTaskCompletionChecked(id int, completed bool) (*model.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	for i := range tasks {
		if tasks[i].ID == id {
			if completed && hasIncompletePrerequisiteLocked(tasks[i]) {
				return nil, ErrTaskBlocked
			}
			if !completed && hasCompletedDependentLocked(id, nil) {
				return nil, ErrTaskBlocked
			}
			tasks[i].Completed = completed
			tasks[i].UpdatedAt = time.Now()
			updated := tasks[i]
			return &updated, nil
		}
	}
	return nil, ErrTaskNotFound
}

func hasIncompletePrerequisiteLocked(task model.Task) bool {
	return hasIncompleteDependencyIDsLocked(task.DependsOn)
}

func hasIncompleteDependencyIDsLocked(dependencyIDs []int) bool {
	for _, dependencyID := range dependencyIDs {
		for _, dependency := range tasks {
			if dependency.ID == dependencyID && !dependency.Completed {
				return true
			}
		}
	}
	return false
}

// BulkSetTaskCompletion updates several tasks atomically.
func BulkSetTaskCompletion(ids []int, completed bool) (BulkUpdateResult, bool) {
	result, err := BulkSetTaskCompletionChecked(ids, completed)
	return result, err == nil
}

// BulkSetTaskCompletionChecked validates task existence and prerequisite state before mutation.
func BulkSetTaskCompletionChecked(ids []int, completed bool) (BulkUpdateResult, error) {
	mu.Lock()
	defer mu.Unlock()

	positions := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		position := -1
		for i := range tasks {
			if tasks[i].ID == id {
				position = i
				break
			}
		}
		if position == -1 {
			return BulkUpdateResult{}, ErrTaskNotFound
		}
		positions = append(positions, position)
	}
	if completed {
		selected := make(map[int]struct{}, len(positions))
		for _, position := range positions {
			selected[tasks[position].ID] = struct{}{}
		}
		for _, position := range positions {
			for _, dependencyID := range tasks[position].DependsOn {
				if _, included := selected[dependencyID]; included {
					continue
				}
				for _, dependency := range tasks {
					if dependency.ID == dependencyID && !dependency.Completed {
						return BulkUpdateResult{}, ErrTaskBlocked
					}
				}
			}
		}
	} else {
		selected := make(map[int]struct{}, len(positions))
		for _, position := range positions {
			selected[tasks[position].ID] = struct{}{}
		}
		for _, position := range positions {
			if hasCompletedDependentLocked(tasks[position].ID, selected) {
				return BulkUpdateResult{}, ErrTaskBlocked
			}
		}
	}

	updated := make([]model.Task, 0, len(positions))
	for _, position := range positions {
		tasks[position].Completed = completed
		tasks[position].UpdatedAt = time.Now()
		updated = append(updated, tasks[position])
	}
	return BulkUpdateResult{Tasks: updated, Updated: len(updated)}, nil
}

// BulkDeleteTasks removes all requested tasks only when every ID exists.
func BulkDeleteTasks(ids []int) (BulkDeleteResult, bool) {
	mu.Lock()
	defer mu.Unlock()

	requested := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		requested[id] = struct{}{}
	}
	for id := range requested {
		found := false
		for _, task := range tasks {
			if task.ID == id {
				found = true
				break
			}
		}
		if !found {
			return BulkDeleteResult{}, false
		}
	}
	for _, task := range tasks {
		if _, deletingDependent := requested[task.ID]; deletingDependent {
			continue
		}
		for _, dependencyID := range task.DependsOn {
			if _, deletingPrerequisite := requested[dependencyID]; deletingPrerequisite {
				return BulkDeleteResult{}, false
			}
		}
	}

	deleted := make([]int, 0, len(requested))
	remaining := make([]model.Task, 0, len(tasks)-len(requested))
	for _, task := range tasks {
		if _, ok := requested[task.ID]; ok {
			deleted = append(deleted, task.ID)
			continue
		}
		remaining = append(remaining, task)
	}
	tasks = remaining
	return BulkDeleteResult{Deleted: deleted, Count: len(deleted)}, true
}

// BulkSetTaskPriority updates priority only when every requested ID exists.
func BulkSetTaskPriority(ids []int, priority string) (BulkUpdateResult, bool) {
	mu.Lock()
	defer mu.Unlock()

	positions := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		position := -1
		for i := range tasks {
			if tasks[i].ID == id {
				position = i
				break
			}
		}
		if position < 0 {
			return BulkUpdateResult{}, false
		}
		positions = append(positions, position)
	}

	updated := make([]model.Task, 0, len(positions))
	for _, position := range positions {
		tasks[position].Priority = priority
		tasks[position].UpdatedAt = time.Now()
		updated = append(updated, tasks[position])
	}
	return BulkUpdateResult{Tasks: updated, Updated: len(updated)}, true
}
