package service

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

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
	ErrChecklistItemNotFound   = errors.New("checklist item not found")
	ErrChecklistTextInvalid    = errors.New("checklist item text is required")
	ErrChecklistLimitReached   = errors.New("task checklist item limit reached")
	ErrChecklistOrderInvalid   = errors.New("checklist order must contain every item exactly once")
	ErrTimerAlreadyRunning     = errors.New("a timer is already running")
	ErrNoActiveTimer           = errors.New("no timer is running for this task")
	ErrTimeEntryNotFound       = errors.New("time entry not found")
	ErrTimeEntryInvalid        = errors.New("time entry is invalid")
	ErrTimeEntryOverlap        = errors.New("time entry overlaps existing tracked time")
	ErrTaskEstimateInvalid     = errors.New("task estimate must be non-negative")
	ErrRecurrenceInvalid       = errors.New("recurrence rule is invalid")
	ErrRecurrenceNotFound      = errors.New("task is not recurring")
)

var activityActions = map[string]struct{}{
	"created": {}, "updated": {}, "completed": {}, "reopened": {}, "deleted": {},
	"dependency_added": {}, "dependency_removed": {}, "dependencies_updated": {},
	"checklist_item_added": {}, "checklist_item_updated": {}, "checklist_item_completed": {},
	"checklist_item_reopened": {}, "checklist_item_deleted": {}, "checklist_reordered": {}, "estimate_updated": {},
	"timer_started": {}, "timer_stopped": {}, "time_logged": {}, "time_entry_deleted": {}, "recurrence_created": {},
}

// IsActivityAction reports whether an action is part of the activity event contract.
func IsActivityAction(action string) bool {
	_, ok := activityActions[action]
	return ok
}

func validateRecurrenceRule(rule *model.RecurrenceRule, dueDate time.Time) error {
	if rule == nil {
		return nil
	}
	if rule.Frequency != "daily" && rule.Frequency != "weekly" && rule.Frequency != "monthly" {
		return ErrRecurrenceInvalid
	}
	if rule.Interval < 1 || rule.Interval > 365 {
		return ErrRecurrenceInvalid
	}
	if rule.Until != nil && rule.Until.Before(dueDate) {
		return ErrRecurrenceInvalid
	}
	return nil
}

// NextRecurrenceDate calculates the next calendar occurrence for a validated rule.
func NextRecurrenceDate(dueDate time.Time, rule *model.RecurrenceRule) (time.Time, error) {
	if err := validateRecurrenceRule(rule, dueDate); err != nil || rule == nil {
		return time.Time{}, ErrRecurrenceInvalid
	}
	switch rule.Frequency {
	case "daily":
		return dueDate.AddDate(0, 0, rule.Interval), nil
	case "weekly":
		return dueDate.AddDate(0, 0, 7*rule.Interval), nil
	case "monthly":
		totalMonths := dueDate.Year()*12 + int(dueDate.Month()) - 1 + rule.Interval
		year := totalMonths / 12
		month := time.Month(totalMonths%12 + 1)
		lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, dueDate.Location()).Day()
		day := dueDate.Day()
		if day > lastDay {
			day = lastDay
		}
		return time.Date(year, month, day, dueDate.Hour(), dueDate.Minute(), dueDate.Second(), dueDate.Nanosecond(), dueDate.Location()), nil
	default:
		return time.Time{}, ErrRecurrenceInvalid
	}
}

// In-memory storage (we'll use database later)
var tasks []model.Task
var activities []model.Activity
var timeEntries []model.TimeEntry
var nextID int = 1
var nextChecklistItemID int = 1
var nextActivityID int = 1
var nextTimeEntryID int = 1
var mu sync.RWMutex

const maxActivityEvents = 10000

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
	activities = nil
	nextActivityID = 1
	timeEntries = nil
	nextTimeEntryID = 1
}

func recordActivityLocked(taskID int, action, summary string) {
	activities = append(activities, model.Activity{
		ID:         nextActivityID,
		TaskID:     taskID,
		Action:     action,
		Summary:    summary,
		OccurredAt: time.Now().UTC(),
	})
	nextActivityID++
	if len(activities) > maxActivityEvents {
		activities = append([]model.Activity(nil), activities[len(activities)-maxActivityEvents:]...)
	}
}

// GetActivityLog returns a snapshot of all recorded task activity.
func GetActivityLog() []model.Activity {
	mu.RLock()
	defer mu.RUnlock()
	return append([]model.Activity{}, activities...)
}

// GetTaskTimeEntries returns a snapshot of a task's recorded time.
func GetTaskTimeEntries(taskID int) ([]model.TimeEntry, bool) {
	mu.RLock()
	defer mu.RUnlock()
	if findTaskPositionLocked(taskID) < 0 {
		return nil, false
	}
	entries := make([]model.TimeEntry, 0)
	now := time.Now()
	for _, entry := range timeEntries {
		if entry.TaskID != taskID {
			continue
		}
		entries = append(entries, currentTimeEntry(entry, now))
	}
	return entries, true
}

// GetActiveTimer returns the currently running timer, if one exists.
func GetActiveTimer() *model.TimeEntry {
	mu.RLock()
	defer mu.RUnlock()
	now := time.Now()
	for _, entry := range timeEntries {
		if entry.EndedAt == nil {
			active := currentTimeEntry(entry, now)
			return &active
		}
	}
	return nil
}

func currentTimeEntry(entry model.TimeEntry, now time.Time) model.TimeEntry {
	if entry.EndedAt == nil {
		entry.DurationSeconds = int64(now.Sub(entry.StartedAt).Seconds())
	}
	return entry
}

// StartTaskTimer starts the single active timer allowed by the service.
func StartTaskTimer(taskID int, note string) (model.TimeEntry, error) {
	mu.Lock()
	defer mu.Unlock()
	if len([]rune(strings.TrimSpace(note))) > 250 {
		return model.TimeEntry{}, ErrTimeEntryInvalid
	}
	taskPosition := findTaskPositionLocked(taskID)
	if taskPosition < 0 {
		return model.TimeEntry{}, ErrTaskNotFound
	}
	for _, entry := range timeEntries {
		if entry.EndedAt == nil {
			return model.TimeEntry{}, ErrTimerAlreadyRunning
		}
	}
	now := time.Now().UTC()
	entry := model.TimeEntry{ID: nextTimeEntryID, TaskID: taskID, StartedAt: now, Note: strings.TrimSpace(note)}
	nextTimeEntryID++
	timeEntries = append(timeEntries, entry)
	tasks[taskPosition].UpdatedAt = now
	recordActivityLocked(taskID, "timer_started", "Started timer for task: "+tasks[taskPosition].Title)
	return currentTimeEntry(entry, now), nil
}

// StopTaskTimer closes the active timer belonging to taskID.
func StopTaskTimer(taskID int) (model.TimeEntry, error) {
	mu.Lock()
	defer mu.Unlock()
	if findTaskPositionLocked(taskID) < 0 {
		return model.TimeEntry{}, ErrTaskNotFound
	}
	for i := range timeEntries {
		if timeEntries[i].TaskID != taskID || timeEntries[i].EndedAt != nil {
			continue
		}
		endedAt := time.Now().UTC()
		timeEntries[i].EndedAt = &endedAt
		timeEntries[i].DurationSeconds = int64(endedAt.Sub(timeEntries[i].StartedAt).Seconds())
		tasks[findTaskPositionLocked(taskID)].UpdatedAt = endedAt
		recordActivityLocked(taskID, "timer_stopped", "Stopped timer after "+strconv.FormatInt(timeEntries[i].DurationSeconds, 10)+" seconds")
		return timeEntries[i], nil
	}
	return model.TimeEntry{}, ErrNoActiveTimer
}

// GetTaskTimeSummary compares a task's estimate with tracked actual time.
func GetTaskTimeSummary(taskID int) (TaskTimeSummary, bool) {
	mu.RLock()
	defer mu.RUnlock()
	taskPosition := findTaskPositionLocked(taskID)
	if taskPosition < 0 {
		return TaskTimeSummary{}, false
	}
	now := time.Now()
	summary := TaskTimeSummary{TaskID: taskID, EstimatedSeconds: int64(tasks[taskPosition].EstimateMinutes) * 60}
	for _, entry := range timeEntries {
		if entry.TaskID == taskID {
			summary.ActualSeconds += currentTimeEntry(entry, now).DurationSeconds
		}
	}
	summary.VarianceSeconds = summary.ActualSeconds - summary.EstimatedSeconds
	return summary, true
}

// GetTimeReport aggregates tracked time over an optional task and time window.
func GetTimeReport(taskID *int, from, to *time.Time) TimeReport {
	mu.RLock()
	defer mu.RUnlock()
	now := time.Now()
	byTask := make(map[int]int64)
	report := TimeReport{ByTask: make([]TaskTimeTotal, 0)}
	for _, entry := range timeEntries {
		if taskID != nil && entry.TaskID != *taskID {
			continue
		}
		start, end := entry.StartedAt, now
		if entry.EndedAt != nil {
			end = *entry.EndedAt
		}
		if from != nil && start.Before(*from) {
			start = *from
		}
		if to != nil && end.After(*to) {
			end = *to
		}
		if !end.After(start) {
			continue
		}
		if entry.EndedAt == nil {
			report.ActiveCount++
		}
		seconds := int64(end.Sub(start).Seconds())
		report.TotalSeconds += seconds
		byTask[entry.TaskID] += seconds
		if entry.EndedAt != nil {
			report.EntryCount++
		}
	}
	for id, seconds := range byTask {
		report.ByTask = append(report.ByTask, TaskTimeTotal{TaskID: id, TotalSeconds: seconds})
	}
	sort.Slice(report.ByTask, func(i, j int) bool { return report.ByTask[i].TaskID < report.ByTask[j].TaskID })
	return report
}

// AddManualTimeEntry records a completed time interval supplied by the caller.
func AddManualTimeEntry(taskID int, startedAt, endedAt time.Time, note string) (model.TimeEntry, error) {
	mu.Lock()
	defer mu.Unlock()
	if findTaskPositionLocked(taskID) < 0 {
		return model.TimeEntry{}, ErrTaskNotFound
	}
	note = strings.TrimSpace(note)
	if !endedAt.After(startedAt) || len([]rune(note)) > 250 {
		return model.TimeEntry{}, ErrTimeEntryInvalid
	}
	startedAt = startedAt.UTC()
	endedAt = endedAt.UTC()
	now := time.Now()
	for _, existing := range timeEntries {
		existingEnd := existing.EndedAt
		if existingEnd == nil {
			existingEnd = &now
		}
		if startedAt.Before(*existingEnd) && endedAt.After(existing.StartedAt) {
			return model.TimeEntry{}, ErrTimeEntryOverlap
		}
	}
	entry := model.TimeEntry{
		ID:              nextTimeEntryID,
		TaskID:          taskID,
		StartedAt:       startedAt,
		EndedAt:         &endedAt,
		DurationSeconds: int64(endedAt.Sub(startedAt).Seconds()),
		Note:            note,
	}
	nextTimeEntryID++
	timeEntries = append(timeEntries, entry)
	taskPosition := findTaskPositionLocked(taskID)
	tasks[taskPosition].UpdatedAt = now.UTC()
	recordActivityLocked(taskID, "time_logged", "Logged "+strconv.FormatInt(entry.DurationSeconds, 10)+" seconds")
	return entry, nil
}

// DeleteTimeEntry removes a completed time entry while retaining an audit event.
func DeleteTimeEntry(taskID, entryID int) error {
	mu.Lock()
	defer mu.Unlock()
	if findTaskPositionLocked(taskID) < 0 {
		return ErrTaskNotFound
	}
	for index := range timeEntries {
		entry := timeEntries[index]
		if entry.TaskID != taskID || entry.ID != entryID {
			continue
		}
		if entry.EndedAt == nil {
			return ErrTimerAlreadyRunning
		}
		timeEntries = append(timeEntries[:index], timeEntries[index+1:]...)
		recordActivityLocked(taskID, "time_entry_deleted", "Deleted time entry of "+strconv.FormatInt(entry.DurationSeconds, 10)+" seconds")
		return nil
	}
	return ErrTimeEntryNotFound
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

type ChecklistProgress struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Remaining int `json:"remaining"`
	Percent   int `json:"percent"`
}

type TaskTimeSummary struct {
	TaskID           int   `json:"taskId"`
	EstimatedSeconds int64 `json:"estimatedSeconds"`
	ActualSeconds    int64 `json:"actualSeconds"`
	VarianceSeconds  int64 `json:"varianceSeconds"`
}

type TaskTimeTotal struct {
	TaskID       int   `json:"taskId"`
	TotalSeconds int64 `json:"totalSeconds"`
}

type TimeReport struct {
	TotalSeconds int64           `json:"totalSeconds"`
	EntryCount   int             `json:"entryCount"`
	ActiveCount  int             `json:"activeCount"`
	ByTask       []TaskTimeTotal `json:"byTask"`
}

type ActivityPage struct {
	Activities []model.Activity `json:"activities"`
	Total      int              `json:"total"`
	Offset     int              `json:"offset"`
	Limit      int              `json:"limit"`
}

// GetAllTasks returns the first page of all tasks.
func GetAllTasks() []model.Task {
	return GetTasks(nil, nil, nil, "", nil, nil, nil, nil, 0, 20, "id", false)
}

// GetActivities returns a filtered page ordered from newest to oldest.
func GetActivities(taskID *int, action string, from, to *time.Time, offset, limit int) ActivityPage {
	mu.RLock()
	defer mu.RUnlock()

	filtered := make([]model.Activity, 0, len(activities))
	for _, activity := range activities {
		if taskID != nil && activity.TaskID != *taskID {
			continue
		}
		if action != "" && activity.Action != action {
			continue
		}
		if from != nil && activity.OccurredAt.Before(*from) {
			continue
		}
		if to != nil && activity.OccurredAt.After(*to) {
			continue
		}
		filtered = append(filtered, activity)
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].OccurredAt.Equal(filtered[j].OccurredAt) {
			return filtered[i].ID > filtered[j].ID
		}
		return filtered[i].OccurredAt.After(filtered[j].OccurredAt)
	})
	total := len(filtered)
	if offset >= total {
		return ActivityPage{Activities: []model.Activity{}, Total: total, Offset: offset, Limit: limit}
	}
	end := offset + limit
	if end < offset || end > total {
		end = total
	}
	return ActivityPage{Activities: filtered[offset:end], Total: total, Offset: offset, Limit: limit}
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
func GetTasks(completed, overdue, recurring *bool, search string, priority *string, tag *string, dueAfter, dueBefore *time.Time, offset, limit int, sortBy string, descending bool) []model.Task {
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
		if recurring != nil && (task.Recurrence != nil) != *recurring {
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
			matchesSearch = strings.Contains(strings.ToLower(task.Description), search)
		}
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
		result = append(result, cloneTask(task))
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

func cloneTask(task model.Task) model.Task {
	task.Tags = append([]string(nil), task.Tags...)
	task.DependsOn = append([]int(nil), task.DependsOn...)
	task.Checklist = append([]model.ChecklistItem(nil), task.Checklist...)
	if task.Recurrence != nil {
		recurrence := *task.Recurrence
		if recurrence.Until != nil {
			until := *recurrence.Until
			recurrence.Until = &until
		}
		task.Recurrence = &recurrence
	}
	return task
}

// GetTaskByID returns a task by ID
func GetTaskByID(id int) *model.Task {
	mu.RLock()
	defer mu.RUnlock()

	for i, task := range tasks {
		if task.ID == id {
			result := cloneTask(tasks[i])
			return &result
		}
	}
	return nil
}

// UpdateTaskEstimate changes the expected task duration in minutes.
func UpdateTaskEstimate(taskID, estimateMinutes int) (*model.Task, error) {
	mu.Lock()
	defer mu.Unlock()
	if estimateMinutes < 0 {
		return nil, ErrTaskEstimateInvalid
	}
	for i := range tasks {
		if tasks[i].ID != taskID {
			continue
		}
		if tasks[i].EstimateMinutes != estimateMinutes {
			previous := tasks[i].EstimateMinutes
			tasks[i].EstimateMinutes = estimateMinutes
			tasks[i].UpdatedAt = time.Now()
			recordActivityLocked(taskID, "estimate_updated", "Updated estimate from "+strconv.Itoa(previous)+" to "+strconv.Itoa(estimateMinutes)+" minutes")
		}
		updated := tasks[i]
		return &updated, nil
	}
	return nil, ErrTaskNotFound
}

// GetTaskChecklist returns a copy of a task's checklist.
func GetTaskChecklist(taskID int) ([]model.ChecklistItem, bool) {
	mu.RLock()
	defer mu.RUnlock()

	for _, task := range tasks {
		if task.ID == taskID {
			items := append([]model.ChecklistItem{}, task.Checklist...)
			return items, true
		}
	}
	return nil, false
}

// GetChecklistProgress returns aggregate completion progress for a task checklist.
func GetChecklistProgress(taskID int) (ChecklistProgress, bool) {
	mu.RLock()
	defer mu.RUnlock()

	for _, task := range tasks {
		if task.ID != taskID {
			continue
		}
		progress := ChecklistProgress{Total: len(task.Checklist)}
		for _, item := range task.Checklist {
			if item.Completed {
				progress.Completed++
			}
		}
		progress.Remaining = progress.Total - progress.Completed
		if progress.Total > 0 {
			progress.Percent = progress.Completed * 100 / progress.Total
		}
		return progress, true
	}
	return ChecklistProgress{}, false
}

// AddChecklistItem appends a new checklist item to a task.
func AddChecklistItem(taskID int, text string) (model.ChecklistItem, error) {
	mu.Lock()
	defer mu.Unlock()

	text = strings.TrimSpace(text)
	if text == "" || utf8.RuneCountInString(text) > 250 {
		return model.ChecklistItem{}, ErrChecklistTextInvalid
	}
	for i := range tasks {
		if tasks[i].ID != taskID {
			continue
		}
		if len(tasks[i].Checklist) >= 100 {
			return model.ChecklistItem{}, ErrChecklistLimitReached
		}
		item := model.ChecklistItem{ID: nextChecklistItemID, Text: text}
		nextChecklistItemID++
		tasks[i].Checklist = append(tasks[i].Checklist, item)
		tasks[i].UpdatedAt = time.Now()
		recordActivityLocked(taskID, "checklist_item_added", "Added checklist item: "+item.Text)
		return item, nil
	}
	return model.ChecklistItem{}, ErrTaskNotFound
}

// SetChecklistItemCompletion updates one checklist item's completion state.
func SetChecklistItemCompletion(taskID, itemID int, completed bool) (model.ChecklistItem, error) {
	mu.Lock()
	defer mu.Unlock()

	for taskIndex := range tasks {
		if tasks[taskIndex].ID != taskID {
			continue
		}
		for itemIndex := range tasks[taskIndex].Checklist {
			if tasks[taskIndex].Checklist[itemIndex].ID == itemID {
				wasCompleted := tasks[taskIndex].Checklist[itemIndex].Completed
				tasks[taskIndex].Checklist[itemIndex].Completed = completed
				tasks[taskIndex].UpdatedAt = time.Now()
				if wasCompleted != completed {
					recordChecklistCompletionActivityLocked(tasks[taskIndex].ID, tasks[taskIndex].Checklist[itemIndex])
				}
				return tasks[taskIndex].Checklist[itemIndex], nil
			}
		}
		return model.ChecklistItem{}, ErrChecklistItemNotFound
	}
	return model.ChecklistItem{}, ErrTaskNotFound
}

// UpdateChecklistItemText changes an item's text without changing its identity or completion state.
func UpdateChecklistItemText(taskID, itemID int, text string) (model.ChecklistItem, error) {
	mu.Lock()
	defer mu.Unlock()

	text = strings.TrimSpace(text)
	if text == "" || utf8.RuneCountInString(text) > 250 {
		return model.ChecklistItem{}, ErrChecklistTextInvalid
	}
	for taskIndex := range tasks {
		if tasks[taskIndex].ID != taskID {
			continue
		}
		for itemIndex := range tasks[taskIndex].Checklist {
			item := &tasks[taskIndex].Checklist[itemIndex]
			if item.ID == itemID {
				oldText := item.Text
				item.Text = text
				tasks[taskIndex].UpdatedAt = time.Now()
				if oldText != text {
					recordActivityLocked(taskID, "checklist_item_updated", "Updated checklist item: "+text)
				}
				return *item, nil
			}
		}
		return model.ChecklistItem{}, ErrChecklistItemNotFound
	}
	return model.ChecklistItem{}, ErrTaskNotFound
}

// UpdateChecklistItem atomically applies provided text and completion changes.
func UpdateChecklistItem(taskID, itemID int, text *string, completed *bool) (model.ChecklistItem, error) {
	mu.Lock()
	defer mu.Unlock()

	if text != nil {
		trimmed := strings.TrimSpace(*text)
		if trimmed == "" || utf8.RuneCountInString(trimmed) > 250 {
			return model.ChecklistItem{}, ErrChecklistTextInvalid
		}
		text = &trimmed
	}
	for taskIndex := range tasks {
		if tasks[taskIndex].ID != taskID {
			continue
		}
		for itemIndex := range tasks[taskIndex].Checklist {
			item := &tasks[taskIndex].Checklist[itemIndex]
			if item.ID != itemID {
				continue
			}
			changed := false
			completionChanged := completed != nil && item.Completed != *completed
			if text != nil && item.Text != *text {
				item.Text = *text
				changed = true
			}
			if completed != nil && item.Completed != *completed {
				item.Completed = *completed
				changed = true
			}
			if changed {
				tasks[taskIndex].UpdatedAt = time.Now()
				if completionChanged {
					recordChecklistCompletionActivityLocked(taskID, *item)
				} else {
					recordActivityLocked(taskID, "checklist_item_updated", "Updated checklist item: "+item.Text)
				}
			}
			return *item, nil
		}
		return model.ChecklistItem{}, ErrChecklistItemNotFound
	}
	return model.ChecklistItem{}, ErrTaskNotFound
}

func recordChecklistCompletionActivityLocked(taskID int, item model.ChecklistItem) {
	if item.Completed {
		recordActivityLocked(taskID, "checklist_item_completed", "Completed checklist item: "+item.Text)
		return
	}
	recordActivityLocked(taskID, "checklist_item_reopened", "Reopened checklist item: "+item.Text)
}

// DeleteChecklistItem removes one checklist item from a task.
func DeleteChecklistItem(taskID, itemID int) error {
	mu.Lock()
	defer mu.Unlock()

	for taskIndex := range tasks {
		if tasks[taskIndex].ID != taskID {
			continue
		}
		for itemIndex := range tasks[taskIndex].Checklist {
			if tasks[taskIndex].Checklist[itemIndex].ID == itemID {
				item := tasks[taskIndex].Checklist[itemIndex]
				tasks[taskIndex].Checklist = append(tasks[taskIndex].Checklist[:itemIndex], tasks[taskIndex].Checklist[itemIndex+1:]...)
				tasks[taskIndex].UpdatedAt = time.Now()
				recordActivityLocked(taskID, "checklist_item_deleted", "Deleted checklist item: "+item.Text)
				return nil
			}
		}
		return ErrChecklistItemNotFound
	}
	return ErrTaskNotFound
}

// ReorderChecklist reorders all checklist items using a complete ordered ID list.
func ReorderChecklist(taskID int, orderedIDs []int) ([]model.ChecklistItem, error) {
	mu.Lock()
	defer mu.Unlock()

	for taskIndex := range tasks {
		if tasks[taskIndex].ID != taskID {
			continue
		}
		items := tasks[taskIndex].Checklist
		if len(orderedIDs) != len(items) {
			return nil, ErrChecklistOrderInvalid
		}
		byID := make(map[int]model.ChecklistItem, len(items))
		for _, item := range items {
			byID[item.ID] = item
		}
		ordered := make([]model.ChecklistItem, 0, len(items))
		for _, id := range orderedIDs {
			item, exists := byID[id]
			if !exists {
				return nil, ErrChecklistOrderInvalid
			}
			ordered = append(ordered, item)
			delete(byID, id)
		}
		if len(byID) != 0 {
			return nil, ErrChecklistOrderInvalid
		}
		changed := false
		for index := range items {
			if items[index].ID != ordered[index].ID {
				changed = true
				break
			}
		}
		tasks[taskIndex].Checklist = ordered
		if changed {
			tasks[taskIndex].UpdatedAt = time.Now()
			recordActivityLocked(taskID, "checklist_reordered", "Reordered task checklist")
		}
		return append([]model.ChecklistItem(nil), ordered...), nil
	}
	return nil, ErrTaskNotFound
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
	recordActivityLocked(taskID, "dependency_added", "Added prerequisite task "+strconv.Itoa(dependencyID))
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
			recordActivityLocked(taskID, "dependency_removed", "Removed prerequisite task "+strconv.Itoa(dependencyID))
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
	changed := !equalIntSlices(original, validated)
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
	if changed {
		recordActivityLocked(taskID, "dependencies_updated", "Updated task prerequisites")
	}
	updated := tasks[taskPosition]
	updated.DependsOn = append([]int(nil), updated.DependsOn...)
	return &updated, nil
}

func equalIntSlices(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
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
	if err := validateRecurrenceRule(task.Recurrence, task.DueDate); err != nil {
		return model.Task{}, err
	}

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
	task.Recurrence = cloneRecurrenceRule(task.Recurrence)
	if task.Recurrence != nil {
		task.RecurrenceSeriesID = task.ID
		task.RecurrenceOccurrence = 1
	} else {
		task.RecurrenceSeriesID = 0
		task.RecurrenceOccurrence = 0
	}
	nextID++
	tasks = append(tasks, task)
	recordActivityLocked(task.ID, "created", "Task created: "+task.Title)
	return task, nil
}

func cloneRecurrenceRule(rule *model.RecurrenceRule) *model.RecurrenceRule {
	if rule == nil {
		return nil
	}
	copy := *rule
	if rule.Until != nil {
		until := *rule.Until
		copy.Until = &until
	}
	return &copy
}

func buildNextOccurrence(source model.Task, id, checklistStartID int, dueDate, now time.Time) (model.Task, int) {
	occurrence := source.RecurrenceOccurrence
	if occurrence < 1 {
		occurrence = 1
	}
	seriesID := source.RecurrenceSeriesID
	if seriesID < 1 {
		seriesID = source.ID
	}
	next := source
	next.ID = id
	next.CreatedAt = now
	next.UpdatedAt = now
	next.DueDate = dueDate
	next.Completed = false
	next.DependsOn = append([]int(nil), source.DependsOn...)
	next.Tags = append([]string(nil), source.Tags...)
	next.Recurrence = cloneRecurrenceRule(source.Recurrence)
	next.RecurrenceSeriesID = seriesID
	next.RecurrenceOccurrence = occurrence + 1
	next.Checklist = make([]model.ChecklistItem, len(source.Checklist))
	for index, item := range source.Checklist {
		next.Checklist[index] = model.ChecklistItem{ID: checklistStartID + index, Text: item.Text}
	}
	return next, checklistStartID + len(next.Checklist)
}

// GetTaskOccurrences returns a snapshot of all tasks in the same recurrence series.
func GetTaskOccurrences(taskID int) ([]model.Task, error) {
	mu.RLock()
	defer mu.RUnlock()
	position := findTaskPositionLocked(taskID)
	if position < 0 {
		return nil, ErrTaskNotFound
	}
	seriesID := tasks[position].RecurrenceSeriesID
	if seriesID < 1 {
		return nil, ErrRecurrenceNotFound
	}
	occurrences := make([]model.Task, 0)
	for _, task := range tasks {
		if task.RecurrenceSeriesID != seriesID {
			continue
		}
		copy := task
		copy.Tags = append([]string(nil), task.Tags...)
		copy.DependsOn = append([]int(nil), task.DependsOn...)
		copy.Checklist = append([]model.ChecklistItem(nil), task.Checklist...)
		copy.Recurrence = cloneRecurrenceRule(task.Recurrence)
		occurrences = append(occurrences, copy)
	}
	sort.Slice(occurrences, func(i, j int) bool {
		return occurrences[i].RecurrenceOccurrence < occurrences[j].RecurrenceOccurrence
	})
	return occurrences, nil
}

// GetNextOccurrenceDueDate previews the next date without creating an occurrence.
func GetNextOccurrenceDueDate(taskID int) (time.Time, error) {
	mu.RLock()
	defer mu.RUnlock()
	position := findTaskPositionLocked(taskID)
	if position < 0 {
		return time.Time{}, ErrTaskNotFound
	}
	task := tasks[position]
	if task.Recurrence == nil {
		return time.Time{}, ErrRecurrenceNotFound
	}
	nextDue, err := NextRecurrenceDate(task.DueDate, task.Recurrence)
	if err != nil || (task.Recurrence.Until != nil && nextDue.After(*task.Recurrence.Until)) {
		return time.Time{}, ErrRecurrenceNotFound
	}
	return nextDue, nil
}

func spawnNextOccurrenceLocked(source model.Task) *model.Task {
	if source.Recurrence == nil {
		return nil
	}
	nextDue, err := NextRecurrenceDate(source.DueDate, source.Recurrence)
	if err != nil || (source.Recurrence.Until != nil && nextDue.After(*source.Recurrence.Until)) {
		return nil
	}
	now := time.Now()
	next, nextChecklistItemIDValue := buildNextOccurrence(source, nextID, nextChecklistItemID, nextDue, now)
	nextChecklistItemID = nextChecklistItemIDValue
	nextID++
	tasks = append(tasks, next)
	recordActivityLocked(source.ID, "recurrence_created", "Created next occurrence "+strconv.Itoa(next.ID)+" due "+nextDue.Format(time.RFC3339))
	return &next
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
			if err := validateRecurrenceRule(task.Recurrence, task.DueDate); err != nil {
				return nil, err
			}
			if task.Completed && hasIncompleteDependencyIDsLocked(dependencyIDs) {
				return nil, ErrTaskBlocked
			}
			if !task.Completed && hasCompletedDependentLocked(id, nil) {
				return nil, ErrTaskBlocked
			}
			oldTitle := tasks[i].Title
			previousSeriesID, previousOccurrence := tasks[i].RecurrenceSeriesID, tasks[i].RecurrenceOccurrence
			task.CreatedAt = tasks[i].CreatedAt
			task.UpdatedAt = time.Now()
			task.Tags = normalizeTags(task.Tags)
			if task.Priority == "" {
				task.Priority = "medium"
			}
			task.ID = id
			task.DependsOn = append([]int(nil), dependencyIDs...)
			task.Recurrence = cloneRecurrenceRule(task.Recurrence)
			if task.Recurrence != nil {
				if previousSeriesID > 0 {
					task.RecurrenceSeriesID = previousSeriesID
					task.RecurrenceOccurrence = previousOccurrence
				} else {
					task.RecurrenceSeriesID = id
					task.RecurrenceOccurrence = 1
				}
			} else if previousSeriesID > 0 {
				task.RecurrenceSeriesID = previousSeriesID
				task.RecurrenceOccurrence = previousOccurrence
			} else {
				task.RecurrenceSeriesID = 0
				task.RecurrenceOccurrence = 0
			}
			tasks[i] = task
			if oldTitle != task.Title {
				recordActivityLocked(id, "updated", "Task renamed from "+oldTitle+" to "+task.Title)
			} else {
				recordActivityLocked(id, "updated", "Task updated: "+task.Title)
			}
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
			recordActivityLocked(id, "deleted", "Task deleted: "+tasks[i].Title)
			tasks = append(tasks[:i], tasks[i+1:]...)
			remainingEntries := timeEntries[:0]
			for _, entry := range timeEntries {
				if entry.TaskID != id {
					remainingEntries = append(remainingEntries, entry)
				}
			}
			timeEntries = remainingEntries
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
			wasCompleted := tasks[i].Completed
			tasks[i].Completed = completed
			tasks[i].UpdatedAt = time.Now()
			if wasCompleted != completed {
				recordCompletionActivityLocked(tasks[i])
				if completed {
					spawnNextOccurrenceLocked(tasks[i])
				}
			}
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
		wasCompleted := tasks[position].Completed
		tasks[position].Completed = completed
		tasks[position].UpdatedAt = time.Now()
		if wasCompleted != completed {
			recordCompletionActivityLocked(tasks[position])
			if completed {
				spawnNextOccurrenceLocked(tasks[position])
			}
		}
		updated = append(updated, tasks[position])
	}
	return BulkUpdateResult{Tasks: updated, Updated: len(updated)}, nil
}

func recordCompletionActivityLocked(task model.Task) {
	if task.Completed {
		recordActivityLocked(task.ID, "completed", "Task completed: "+task.Title)
		return
	}
	recordActivityLocked(task.ID, "reopened", "Task reopened: "+task.Title)
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
			recordActivityLocked(task.ID, "deleted", "Task deleted: "+task.Title)
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
