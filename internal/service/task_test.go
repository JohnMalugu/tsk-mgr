package service

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

func TestDependencyErrorsAreDistinct(t *testing.T) {
	if errors.Is(ErrDependencyCycle, ErrDependencyAlreadyExists) {
		t.Fatal("cycle and duplicate errors must remain distinct")
	}
	if ErrTaskBlocked.Error() != "task has incomplete prerequisites" {
		t.Fatalf("unexpected blocked-task error: %q", ErrTaskBlocked)
	}
}

func TestTaskSummaryIncludesWorkloadAndDueSoonMetrics(t *testing.T) {
	ResetTasks()
	if _, err := UpdateTaskEstimate(1, 30); err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateTaskEstimate(2, 45); err != nil {
		t.Fatal(err)
	}
	startedAt := time.Now().Add(-30 * time.Minute)
	if _, err := AddManualTimeEntry(1, startedAt, startedAt.Add(15*time.Minute), "Focused work"); err != nil {
		t.Fatal(err)
	}
	summary := GetTaskSummary()
	if summary.EstimatedMinutes != 75 || summary.TrackedSeconds != 900 || summary.DueSoon != 1 {
		t.Fatalf("unexpected workload summary: %#v", summary)
	}
}

func TestTaskSummaryCountsStatuses(t *testing.T) {
	ResetTasks()
	if _, err := UpdateTaskWithDependencies(1, model.Task{Title: "Running", DueDate: time.Now().Add(time.Hour), Status: model.TaskStatusInProgress}, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := SetTaskCompletionChecked(2, true); err != nil {
		t.Fatal(err)
	}
	summary := GetTaskSummary()
	if summary.ByStatus[model.TaskStatusTodo] != 0 || summary.ByStatus[model.TaskStatusInProgress] != 1 || summary.ByStatus[model.TaskStatusCompleted] != 1 || summary.ByStatus[model.TaskStatusCanceled] != 0 {
		t.Fatalf("unexpected status counts: %#v", summary.ByStatus)
	}
}

func TestGetTagSummaryCountsTasksAndSortsDeterministically(t *testing.T) {
	ResetTasks()
	CreateTask(model.Task{Title: "Tagged one", Tags: []string{"home", "work"}})
	CreateTask(model.Task{Title: "Tagged two", Tags: []string{"work"}})
	summary := GetTagSummary(0, 20)
	if len(summary.Tags) != 4 || summary.Tags[0].Tag != "home" || summary.Tags[0].TaskCount != 2 || summary.Tags[1].Tag != "work" || summary.Tags[1].TaskCount != 2 {
		t.Fatalf("unexpected tag usage summary: %#v", summary)
	}
	if summary.Tags[2].Tag != "errands" || summary.Tags[3].Tag != "study" {
		t.Fatalf("expected stable alphabetical order for ties, got %#v", summary)
	}
}

func TestTaskStatusStaysCompatibleWithCompletion(t *testing.T) {
	ResetTasks()
	task, err := CreateTaskWithDependencies(model.Task{Title: "Status task", Status: model.TaskStatusInProgress}, nil)
	if err != nil || task.Status != model.TaskStatusInProgress || task.Completed {
		t.Fatalf("unexpected new task status: %#v err=%v", task, err)
	}
	completed, err := SetTaskCompletionChecked(task.ID, true)
	if err != nil || completed.Status != model.TaskStatusCompleted || !completed.Completed {
		t.Fatalf("unexpected completed task state: %#v err=%v", completed, err)
	}
	reopened, err := SetTaskCompletionChecked(task.ID, false)
	if err != nil || reopened.Status != model.TaskStatusTodo || reopened.Completed {
		t.Fatalf("unexpected reopened task state: %#v err=%v", reopened, err)
	}
}

func TestTaskCommentsValidateAndRecordChronologically(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskComment(1, "   "); !errors.Is(err, ErrCommentInvalid) {
		t.Fatalf("expected empty comment rejection, got %v", err)
	}
	first, err := AddTaskComment(1, "  First note  ")
	if err != nil || first.Body != "First note" {
		t.Fatalf("unexpected first comment: %#v err=%v", first, err)
	}
	second, err := AddTaskComment(1, "Second note")
	if err != nil || second.ID <= first.ID {
		t.Fatalf("unexpected second comment: %#v err=%v", second, err)
	}
	comments, found := GetTaskComments(1, 0, 20)
	if !found || comments.Total != 2 || len(comments.Comments) != 2 || comments.Comments[0].ID != first.ID || comments.Comments[1].ID != second.ID {
		t.Fatalf("unexpected task comments: %#v found=%v", comments, found)
	}
	page, found := GetTaskComments(1, 1, 1)
	if !found || page.Total != 2 || len(page.Comments) != 1 || page.Comments[0].ID != second.ID {
		t.Fatalf("unexpected paginated comments: %#v found=%v", page, found)
	}
	if !IsActivityAction("comment_added") {
		t.Fatal("comment activity must be a supported filter action")
	}
	if _, found := GetTaskComments(999, 0, 20); found {
		t.Fatal("expected missing task to be reported")
	}
}

func TestDeleteTaskCommentRecordsActivity(t *testing.T) {
	ResetTasks()
	comment, err := AddTaskComment(1, "Remove this note")
	if err != nil {
		t.Fatal(err)
	}
	if err := DeleteTaskComment(1, comment.ID); err != nil {
		t.Fatal(err)
	}
	page, found := GetTaskComments(1, 0, 20)
	if !found || page.Total != 0 {
		t.Fatalf("expected deleted comment to be absent: %#v", page)
	}
	events := GetActivities(nil, "comment_deleted", nil, nil, 0, 20)
	if events.Total != 1 || events.Activities[0].TaskID != 1 {
		t.Fatalf("expected deletion audit event, got %#v", events)
	}
	if err := DeleteTaskComment(1, comment.ID); !errors.Is(err, ErrCommentNotFound) {
		t.Fatalf("expected missing-comment error, got %v", err)
	}
	if err := DeleteTaskComment(999, comment.ID); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected missing-task error, got %v", err)
	}
}

func TestUpdateTaskCommentChangesBodyAndRecordsActivity(t *testing.T) {
	ResetTasks()
	comment, err := AddTaskComment(1, "Initial note")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := UpdateTaskComment(1, comment.ID, "  Revised note  ")
	if err != nil || updated.Body != "Revised note" || !updated.UpdatedAt.After(comment.CreatedAt) {
		t.Fatalf("unexpected updated comment: %#v err=%v", updated, err)
	}
	page, found := GetTaskComments(1, 0, 20)
	if !found || page.Total != 1 || page.Comments[0].Body != "Revised note" {
		t.Fatalf("unexpected persisted comment: %#v", page)
	}
	if events := GetActivities(nil, "comment_updated", nil, nil, 0, 20); events.Total != 1 {
		t.Fatalf("expected one comment update event, got %#v", events)
	}
}

func TestCompletedStatusSetsLegacyCompletionOnCreateAndUpdate(t *testing.T) {
	ResetTasks()
	created := CreateTask(model.Task{Title: "Completed via status", Status: model.TaskStatusCompleted})
	if !created.Completed || created.Status != model.TaskStatusCompleted {
		t.Fatalf("status not reflected in legacy completion field: %#v", created)
	}
	existing := GetTaskByID(2)
	if existing == nil {
		t.Fatal("expected seeded task")
	}
	existing.Status = model.TaskStatusCompleted
	existing.Completed = false
	updated := UpdateTask(existing.ID, *existing)
	if updated == nil || !updated.Completed || updated.Status != model.TaskStatusCompleted {
		t.Fatalf("updated status not reflected in completion field: %#v", updated)
	}
}

func TestValidateRecurrenceRuleBoundsAndDates(t *testing.T) {
	due := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	valid := &model.RecurrenceRule{Frequency: "weekly", Interval: 2}
	if err := validateRecurrenceRule(valid, due); err != nil {
		t.Fatalf("expected valid recurrence rule: %v", err)
	}
	for _, rule := range []*model.RecurrenceRule{
		{Frequency: "yearly", Interval: 1},
		{Frequency: "daily", Interval: 0},
		{Frequency: "monthly", Interval: 366},
		{Frequency: "daily", Interval: 1, Until: timePtr(due.Add(-time.Second))},
	} {
		if err := validateRecurrenceRule(rule, due); !errors.Is(err, ErrRecurrenceInvalid) {
			t.Fatalf("expected invalid rule %#v, got %v", rule, err)
		}
	}
}

func TestNextRecurrenceDateCalendarBoundaries(t *testing.T) {
	monthly := &model.RecurrenceRule{Frequency: "monthly", Interval: 1}
	jan31 := time.Date(2027, time.January, 31, 9, 15, 0, 0, time.UTC)
	if got, _ := NextRecurrenceDate(jan31, monthly); got.Format(time.RFC3339) != "2027-02-28T09:15:00Z" {
		t.Fatalf("expected month-end clamp, got %v", got)
	}
	leapJan31 := time.Date(2028, time.January, 31, 9, 15, 0, 0, time.UTC)
	if got, _ := NextRecurrenceDate(leapJan31, monthly); got.Format(time.RFC3339) != "2028-02-29T09:15:00Z" {
		t.Fatalf("expected leap-year month-end clamp, got %v", got)
	}
	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}
	daily := &model.RecurrenceRule{Frequency: "daily", Interval: 1}
	springDay := time.Date(2026, time.March, 7, 9, 0, 0, 0, location)
	if got, _ := NextRecurrenceDate(springDay, daily); got.Hour() != 9 || got.Day() != 8 || got.Location() != location {
		t.Fatalf("expected local wall time preserved over DST, got %v", got)
	}
}

func timePtr(value time.Time) *time.Time { return &value }

func TestResetTasksClearsTimeEntryState(t *testing.T) {
	timeEntries = []model.TimeEntry{{ID: 8, TaskID: 1}}
	nextTimeEntryID = 9
	ResetTasks()
	if len(timeEntries) != 0 || nextTimeEntryID != 1 {
		t.Fatalf("expected time tracking state reset, entries=%#v nextID=%d", timeEntries, nextTimeEntryID)
	}
}

func TestGetTaskTimeEntriesDistinguishesMissingTask(t *testing.T) {
	ResetTasks()
	entries, found := GetTaskTimeEntries(1)
	if !found || entries == nil || len(entries) != 0 {
		t.Fatalf("expected empty time-entry list, got %#v found=%v", entries, found)
	}
	if _, found := GetTaskTimeEntries(999); found {
		t.Fatal("expected missing task to be reported")
	}
	if GetActiveTimer() != nil {
		t.Fatal("expected no active timer")
	}
}

func TestStartTaskTimerIsExclusive(t *testing.T) {
	ResetTasks()
	started, err := StartTaskTimer(1, "  Focus block  ")
	if err != nil {
		t.Fatal(err)
	}
	if started.ID != 1 || started.Note != "Focus block" || started.EndedAt != nil {
		t.Fatalf("unexpected active timer: %#v", started)
	}
	if _, err := StartTaskTimer(2, "another task"); !errors.Is(err, ErrTimerAlreadyRunning) {
		t.Fatalf("expected active timer conflict, got %v", err)
	}
	if _, err := StartTaskTimer(999, "missing"); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected missing task error, got %v", err)
	}
}

func TestConcurrentTimerStartsHaveSingleWinner(t *testing.T) {
	ResetTasks()
	const attempts = 32
	var wait sync.WaitGroup
	results := make(chan error, attempts)
	for index := 0; index < attempts; index++ {
		wait.Add(1)
		go func(taskID int) {
			defer wait.Done()
			_, err := StartTaskTimer(taskID, "Concurrent")
			results <- err
		}(index%2 + 1)
	}
	wait.Wait()
	close(results)
	winners := 0
	for err := range results {
		if err == nil {
			winners++
		} else if !errors.Is(err, ErrTimerAlreadyRunning) {
			t.Fatalf("unexpected concurrent timer error: %v", err)
		}
	}
	if winners != 1 {
		t.Fatalf("expected exactly one active timer start, got %d", winners)
	}
}

func TestConcurrentCompletionGeneratesOneRecurringOccurrence(t *testing.T) {
	ResetTasks()
	task, err := CreateTaskWithDependencies(model.Task{
		Title: "Recurring task", DueDate: time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC),
		Recurrence: &model.RecurrenceRule{Frequency: "daily", Interval: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	const attempts = 24
	var wait sync.WaitGroup
	errorsFound := make(chan error, attempts)
	for range attempts {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, err := SetTaskCompletionChecked(task.ID, true)
			errorsFound <- err
		}()
	}
	wait.Wait()
	close(errorsFound)
	for err := range errorsFound {
		if err != nil {
			t.Fatalf("concurrent idempotent completion failed: %v", err)
		}
	}
	occurrences, err := GetTaskOccurrences(task.ID)
	if err != nil || len(occurrences) != 2 {
		t.Fatalf("expected exactly two series occurrences, got %#v err=%v", occurrences, err)
	}
}

func TestStopTaskTimerPersistsElapsedTime(t *testing.T) {
	ResetTasks()
	if _, err := StartTaskTimer(1, "Focus"); err != nil {
		t.Fatal(err)
	}
	stopped, err := StopTaskTimer(1)
	if err != nil {
		t.Fatal(err)
	}
	if stopped.EndedAt == nil || stopped.DurationSeconds < 0 {
		t.Fatalf("expected completed time entry, got %#v", stopped)
	}
	if _, err := StopTaskTimer(1); !errors.Is(err, ErrNoActiveTimer) {
		t.Fatalf("expected no-active-timer error, got %v", err)
	}
	if _, err := StopTaskTimer(999); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected missing-task error, got %v", err)
	}
}

func TestTimerTransitionsRecordActivity(t *testing.T) {
	ResetTasks()
	if _, err := StartTaskTimer(1, "Focus"); err != nil {
		t.Fatal(err)
	}
	if _, err := StopTaskTimer(1); err != nil {
		t.Fatal(err)
	}
	events := GetActivities(nil, "", nil, nil, 0, 10)
	if events.Total != 2 || events.Activities[0].Action != "timer_stopped" || events.Activities[1].Action != "timer_started" {
		t.Fatalf("unexpected timer activity: %#v", events)
	}
}

func TestAddManualTimeEntryValidatesIntervalsAndOverlaps(t *testing.T) {
	ResetTasks()
	start := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	entry, err := AddManualTimeEntry(1, start, end, " Planning ")
	if err != nil || entry.DurationSeconds != 3600 || entry.Note != "Planning" || entry.EndedAt == nil {
		t.Fatalf("unexpected manual time entry: %#v err=%v", entry, err)
	}
	if _, err := AddManualTimeEntry(2, start.Add(30*time.Minute), end.Add(time.Hour), "Overlap"); !errors.Is(err, ErrTimeEntryOverlap) {
		t.Fatalf("expected overlap error, got %v", err)
	}
	if _, err := AddManualTimeEntry(1, end, start, "Invalid"); !errors.Is(err, ErrTimeEntryInvalid) {
		t.Fatalf("expected invalid interval error, got %v", err)
	}
}

func TestManualTimeEntryRecordsActivity(t *testing.T) {
	ResetTasks()
	start := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	if _, err := AddManualTimeEntry(1, start, start.Add(20*time.Minute), "Planning"); err != nil {
		t.Fatal(err)
	}
	page := GetActivities(ptrInt(1), "time_logged", nil, nil, 0, 10)
	if page.Total != 1 || len(page.Activities) != 1 || page.Activities[0].Summary != "Logged 1200 seconds" {
		t.Fatalf("unexpected manual time activity: %#v", page)
	}
}

func ptrInt(value int) *int { return &value }

func TestDeleteTimeEntryRejectsActiveTimer(t *testing.T) {
	ResetTasks()
	active, err := StartTaskTimer(1, "Focus")
	if err != nil {
		t.Fatal(err)
	}
	if err := DeleteTimeEntry(1, active.ID); !errors.Is(err, ErrTimerAlreadyRunning) {
		t.Fatalf("expected active timer deletion rejection, got %v", err)
	}
	stopped, err := StopTaskTimer(1)
	if err != nil {
		t.Fatal(err)
	}
	if err := DeleteTimeEntry(1, stopped.ID); err != nil {
		t.Fatal(err)
	}
	entries, _ := GetTaskTimeEntries(1)
	if len(entries) != 0 {
		t.Fatalf("expected deleted entry omitted from task time, got %#v", entries)
	}
}

func TestResetTasksClearsActivityLogAndSequence(t *testing.T) {
	ResetTasks()
	mu.Lock()
	recordActivityLocked(1, "created", "Task created")
	mu.Unlock()
	ResetTasks()
	if events := GetActivityLog(); events == nil || len(events) != 0 {
		t.Fatalf("expected empty activity log after reset, got %#v", events)
	}
	mu.Lock()
	recordActivityLocked(1, "created", "Task created")
	mu.Unlock()
	events := GetActivityLog()
	if len(events) != 1 || events[0].ID != 1 {
		t.Fatalf("expected activity ID sequence to reset, got %#v", events)
	}
}

func TestActivityLogRetainsNewestEventsWithinLimit(t *testing.T) {
	ResetTasks()
	mu.Lock()
	for index := 0; index < maxActivityEvents+2; index++ {
		recordActivityLocked(1, "updated", "event")
	}
	mu.Unlock()
	events := GetActivityLog()
	if len(events) != maxActivityEvents {
		t.Fatalf("expected %d retained events, got %d", maxActivityEvents, len(events))
	}
	if events[0].ID != 3 || events[len(events)-1].ID != maxActivityEvents+2 {
		t.Fatalf("expected newest event window, got IDs %d through %d", events[0].ID, events[len(events)-1].ID)
	}
}

func TestGetActivitiesFiltersAndPaginatesNewestFirst(t *testing.T) {
	ResetTasks()
	CreateTask(model.Task{Title: "First"})
	CreateTask(model.Task{Title: "Second"})
	action := "created"
	page := GetActivities(nil, action, nil, nil, 1, 1)
	if page.Total != 2 || page.Offset != 1 || page.Limit != 1 || len(page.Activities) != 1 {
		t.Fatalf("unexpected activity page metadata: %#v", page)
	}
	if page.Activities[0].Summary != "Task created: First" {
		t.Fatalf("expected second-newest event, got %#v", page.Activities[0])
	}
	taskID := page.Activities[0].TaskID
	filtered := GetActivities(&taskID, action, nil, nil, 0, 10)
	if filtered.Total != 1 || len(filtered.Activities) != 1 {
		t.Fatalf("expected one task-filtered event, got %#v", filtered)
	}
}
func TestChecklistErrorsAreDistinct(t *testing.T) {
	if errors.Is(ErrChecklistItemNotFound, ErrTaskNotFound) {
		t.Fatal("missing checklist item and missing task must be distinct")
	}
	if ErrChecklistOrderInvalid.Error() != "checklist order must contain every item exactly once" {
		t.Fatalf("unexpected checklist order error: %q", ErrChecklistOrderInvalid)
	}
}

func TestGetTaskByIDReturnsCopy(t *testing.T) {
	ResetTasks()
	until := time.Now().Add(72 * time.Hour)
	created := CreateTask(model.Task{
		Title: "copy test", DueDate: time.Now().Add(48 * time.Hour),
		Tags: []string{"original"}, DependsOn: []int{1},
		Checklist:  []model.ChecklistItem{{ID: 7, Text: "original"}},
		Recurrence: &model.RecurrenceRule{Frequency: "daily", Interval: 1, Until: &until},
	})
	task := GetTaskByID(created.ID)
	if task == nil {
		t.Fatal("expected created task")
	}

	task.Title = "locally changed"
	task.Tags[0] = "changed"
	task.DependsOn[0] = 99
	task.Checklist[0].Text = "changed"
	task.Recurrence.Frequency = "weekly"
	task.Recurrence.Until = nil

	storedTask := GetTaskByID(created.ID)
	if storedTask == nil {
		t.Fatal("expected stored task")
	}
	if storedTask.Title != "copy test" || storedTask.Tags[0] != "original" || storedTask.DependsOn[0] != 1 || storedTask.Checklist[0].Text != "original" || storedTask.Recurrence.Frequency != "daily" || storedTask.Recurrence.Until == nil {
		t.Fatalf("lookup result mutated stored task: %#v", storedTask)
	}
	listed := GetAllTasks()
	for i := range listed {
		if listed[i].ID == created.ID {
			listed[i].Tags[0] = "list mutation"
		}
	}
	storedTask = GetTaskByID(created.ID)
	if storedTask.Tags[0] != "original" {
		t.Fatalf("list result mutated stored task tags: %#v", storedTask.Tags)
	}
}

func TestCreateAndUpdateTaskReturnIndependentChecklistSnapshots(t *testing.T) {
	ResetTasks()
	created := CreateTask(model.Task{Title: "Snapshot", Checklist: []model.ChecklistItem{{Text: "Original"}}})
	created.Checklist[0].Text = "Mutated create result"
	stored := GetTaskByID(created.ID)
	if stored == nil || stored.Checklist[0].Text != "Original" {
		t.Fatalf("create result mutated stored checklist: %#v", stored)
	}
	update := *stored
	update.Checklist[0].Text = "Updated"
	updated := UpdateTask(created.ID, update)
	if updated == nil {
		t.Fatal("expected task update")
	}
	updated.Checklist[0].Text = "Mutated update result"
	stored = GetTaskByID(created.ID)
	if stored == nil || stored.Checklist[0].Text != "Updated" {
		t.Fatalf("update result mutated stored checklist: %#v", stored)
	}
}

func TestGetTaskChecklistReturnsCopyAndDistinguishesMissing(t *testing.T) {
	ResetTasks()
	items, found := GetTaskChecklist(1)
	if !found || items == nil || len(items) != 0 {
		t.Fatalf("expected existing task with empty checklist, got %#v, found=%v", items, found)
	}
	if _, found := GetTaskChecklist(999); found {
		t.Fatal("expected missing task to be reported")
	}
}

func TestAddChecklistItemTrimsTextAndAssignsID(t *testing.T) {
	ResetTasks()
	item, err := AddChecklistItem(1, "  buy milk  ")
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != 1 || item.Text != "buy milk" || item.Completed {
		t.Fatalf("unexpected checklist item: %#v", item)
	}
	items, _ := GetTaskChecklist(1)
	if len(items) != 1 || items[0] != item {
		t.Fatalf("expected item stored on task, got %#v", items)
	}
}

func TestAddChecklistItemRecordsActivity(t *testing.T) {
	ResetTasks()
	if _, err := AddChecklistItem(1, "Review release notes"); err != nil {
		t.Fatal(err)
	}
	events := GetActivityLog()
	if len(events) != 1 || events[0].Action != "checklist_item_added" || events[0].Summary != "Added checklist item: Review release notes" {
		t.Fatalf("unexpected checklist activity: %#v", events)
	}
}

func TestSetChecklistItemCompletion(t *testing.T) {
	ResetTasks()
	item, err := AddChecklistItem(1, "draft")
	if err != nil {
		t.Fatal(err)
	}
	updated, err := SetChecklistItemCompletion(1, item.ID, true)
	if err != nil || !updated.Completed {
		t.Fatalf("expected item completion, item=%#v err=%v", updated, err)
	}
	if _, err := SetChecklistItemCompletion(1, 999, true); !errors.Is(err, ErrChecklistItemNotFound) {
		t.Fatalf("expected missing-item error, got %v", err)
	}
}

func TestBulkSetChecklistCompletionIsAtomic(t *testing.T) {
	ResetTasks()
	first, err := AddChecklistItem(1, "First")
	if err != nil {
		t.Fatal(err)
	}
	second, err := AddChecklistItem(1, "Second")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BulkSetChecklistCompletion(1, []int{first.ID, 999}, true); !errors.Is(err, ErrChecklistItemNotFound) {
		t.Fatalf("expected missing item error, got %v", err)
	}
	items, _ := GetTaskChecklist(1)
	if items[0].Completed || items[1].Completed {
		t.Fatalf("rejected batch partially changed checklist: %#v", items)
	}
	updated, err := BulkSetChecklistCompletion(1, []int{first.ID, second.ID, first.ID}, true)
	if err != nil || len(updated) != 2 || !updated[0].Completed || !updated[1].Completed {
		t.Fatalf("unexpected bulk checklist result: %#v err=%v", updated, err)
	}
}

func TestUpdateChecklistItemTextPreservesState(t *testing.T) {
	ResetTasks()
	item, err := AddChecklistItem(1, "draft")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SetChecklistItemCompletion(1, item.ID, true); err != nil {
		t.Fatal(err)
	}
	updated, err := UpdateChecklistItemText(1, item.ID, "  review  ")
	if err != nil || updated.ID != item.ID || updated.Text != "review" || !updated.Completed {
		t.Fatalf("expected text-only update, got %#v err=%v", updated, err)
	}
}

func TestUpdateChecklistItemAppliesFieldsAtomically(t *testing.T) {
	ResetTasks()
	item, err := AddChecklistItem(1, "draft")
	if err != nil {
		t.Fatal(err)
	}
	invalidText := "  "
	completed := true
	if _, err := UpdateChecklistItem(1, item.ID, &invalidText, &completed); !errors.Is(err, ErrChecklistTextInvalid) {
		t.Fatalf("expected text validation error, got %v", err)
	}
	items, _ := GetTaskChecklist(1)
	if items[0].Completed || items[0].Text != "draft" {
		t.Fatalf("invalid update partially mutated item: %#v", items[0])
	}
}

func TestDeleteChecklistItem(t *testing.T) {
	ResetTasks()
	item, err := AddChecklistItem(1, "remove me")
	if err != nil {
		t.Fatal(err)
	}
	if err := DeleteChecklistItem(1, item.ID); err != nil {
		t.Fatal(err)
	}
	items, _ := GetTaskChecklist(1)
	if len(items) != 0 {
		t.Fatalf("expected empty checklist after deletion, got %#v", items)
	}
	if err := DeleteChecklistItem(1, item.ID); !errors.Is(err, ErrChecklistItemNotFound) {
		t.Fatalf("expected missing-item error, got %v", err)
	}
}

func TestChecklistMutationsRecordMeaningfulActivity(t *testing.T) {
	ResetTasks()
	item, err := AddChecklistItem(1, "Draft")
	if err != nil {
		t.Fatal(err)
	}
	completed := true
	if _, err := UpdateChecklistItem(1, item.ID, nil, &completed); err != nil {
		t.Fatal(err)
	}
	if _, err := UpdateChecklistItem(1, item.ID, nil, &completed); err != nil {
		t.Fatal(err)
	}
	if err := DeleteChecklistItem(1, item.ID); err != nil {
		t.Fatal(err)
	}
	events := GetActivityLog()
	if len(events) != 3 || events[1].Action != "checklist_item_completed" || events[2].Action != "checklist_item_deleted" {
		t.Fatalf("expected add/complete/delete events without no-op duplicate, got %#v", events)
	}
}

func TestReorderChecklistRequiresExactItemSet(t *testing.T) {
	ResetTasks()
	first, _ := AddChecklistItem(1, "first")
	second, _ := AddChecklistItem(1, "second")
	ordered, err := ReorderChecklist(1, []int{second.ID, first.ID})
	if err != nil || len(ordered) != 2 || ordered[0].ID != second.ID {
		t.Fatalf("expected reversed order, got %#v err=%v", ordered, err)
	}
	if _, err := ReorderChecklist(1, []int{second.ID, second.ID}); !errors.Is(err, ErrChecklistOrderInvalid) {
		t.Fatalf("expected duplicate ID rejection, got %v", err)
	}
	after, _ := GetTaskChecklist(1)
	if len(after) != 2 || after[0].ID != second.ID || after[1].ID != first.ID {
		t.Fatalf("invalid reorder must preserve current order, got %#v", after)
	}
}

func TestGetChecklistProgress(t *testing.T) {
	ResetTasks()
	if _, err := AddChecklistItem(1, "first"); err != nil {
		t.Fatal(err)
	}
	second, err := AddChecklistItem(1, "second")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SetChecklistItemCompletion(1, second.ID, true); err != nil {
		t.Fatal(err)
	}
	progress, found := GetChecklistProgress(1)
	if !found || progress.Total != 2 || progress.Completed != 1 || progress.Remaining != 1 || progress.Percent != 50 {
		t.Fatalf("unexpected progress: %#v found=%v", progress, found)
	}
	empty, _ := GetChecklistProgress(2)
	if empty.Total != 0 || empty.Percent != 0 {
		t.Fatalf("expected empty checklist progress, got %#v", empty)
	}
}

func TestResetTasksResetsChecklistItemSequence(t *testing.T) {
	nextChecklistItemID = 19
	ResetTasks()
	if nextChecklistItemID != 1 {
		t.Fatalf("expected checklist ID sequence to reset to 1, got %d", nextChecklistItemID)
	}
}

func TestGetTaskDependenciesDistinguishesEmptyFromMissing(t *testing.T) {
	ResetTasks()
	dependencies, found := GetTaskDependencies(1)
	if !found || dependencies == nil || len(dependencies) != 0 {
		t.Fatalf("expected existing task with empty dependency list, got %#v, found=%v", dependencies, found)
	}
	if _, found := GetTaskDependencies(999); found {
		t.Fatal("expected missing task to be reported")
	}
}

func TestGetTaskDependentsReturnsReverseEdges(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	dependents, found := GetTaskDependents(2)
	if !found || len(dependents) != 1 || dependents[0].ID != 1 {
		t.Fatalf("unexpected dependents: %#v found=%v", dependents, found)
	}
	if _, found := GetTaskDependents(999); found {
		t.Fatal("expected missing prerequisite to be reported")
	}
}

func TestAddTaskDependencyRejectsInvalidGraphEdges(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 1); !errors.Is(err, ErrDependencySelfReference) {
		t.Fatalf("expected self-reference error, got %v", err)
	}
	if _, err := AddTaskDependency(1, 999); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected missing-task error, got %v", err)
	}
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatalf("add first edge: %v", err)
	}
	if _, err := AddTaskDependency(1, 2); !errors.Is(err, ErrDependencyAlreadyExists) {
		t.Fatalf("expected duplicate-edge error, got %v", err)
	}
	if _, err := AddTaskDependency(2, 1); !errors.Is(err, ErrDependencyCycle) {
		t.Fatalf("expected cycle error, got %v", err)
	}
	if task := GetTaskByID(2); task == nil || len(task.DependsOn) != 0 {
		t.Fatal("cycle rejection must not mutate the graph")
	}
}

func TestDependencyChangesRecordActivity(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := RemoveTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	events := GetActivityLog()
	if len(events) != 2 || events[0].Action != "dependency_added" || events[1].Action != "dependency_removed" {
		t.Fatalf("unexpected dependency activity: %#v", events)
	}
}

func TestReplaceTaskDependenciesIsAtomicAndCycleSafe(t *testing.T) {
	ResetTasks()
	created := CreateTask(model.Task{Title: "Third task"})
	if _, err := AddTaskDependency(2, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := ReplaceTaskDependencies(1, []int{created.ID, 2}); !errors.Is(err, ErrDependencyCycle) {
		t.Fatalf("expected cycle error, got %v", err)
	}
	task := GetTaskByID(1)
	if task == nil || len(task.DependsOn) != 0 {
		t.Fatalf("expected original dependency set preserved, got %#v", task)
	}
}

func TestCompletedTaskCannotReceiveIncompletePrerequisite(t *testing.T) {
	ResetTasks()
	if _, err := SetTaskCompletionChecked(1, true); err != nil {
		t.Fatal(err)
	}
	if _, err := AddTaskDependency(1, 2); !errors.Is(err, ErrTaskBlocked) {
		t.Fatalf("expected dependency add rejection, got %v", err)
	}
	if _, err := ReplaceTaskDependencies(1, []int{2}); !errors.Is(err, ErrTaskBlocked) {
		t.Fatalf("expected dependency replacement rejection, got %v", err)
	}
	if _, err := CreateTaskWithDependencies(model.Task{Title: "Already complete", Completed: true}, []int{2}); !errors.Is(err, ErrTaskBlocked) {
		t.Fatalf("expected completed-task creation rejection, got %v", err)
	}
	if _, err := UpdateTaskWithDependencies(1, model.Task{Title: "Still complete", Completed: true}, []int{2}); !errors.Is(err, ErrTaskBlocked) {
		t.Fatalf("expected completed-task update rejection, got %v", err)
	}
}

func TestPrerequisiteCannotBeUncompletedWhileDependentIsComplete(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := BulkSetTaskCompletionChecked([]int{1, 2}, true); err != nil {
		t.Fatal(err)
	}
	if _, err := SetTaskCompletionChecked(2, false); !errors.Is(err, ErrTaskBlocked) {
		t.Fatalf("expected prerequisite reset to be rejected, got %v", err)
	}
	if _, err := BulkSetTaskCompletionChecked([]int{2}, false); !errors.Is(err, ErrTaskBlocked) {
		t.Fatalf("expected partial reset batch to be rejected, got %v", err)
	}
	prerequisite := GetTaskByID(2)
	if prerequisite == nil {
		t.Fatal("expected prerequisite task to exist")
	}
	if _, err := UpdateTaskWithDependencies(2, model.Task{Title: prerequisite.Title, DueDate: prerequisite.DueDate, Completed: false}, prerequisite.DependsOn); !errors.Is(err, ErrTaskBlocked) {
		t.Fatalf("expected full update reset to be rejected, got %v", err)
	}
	if _, err := BulkSetTaskCompletionChecked([]int{1, 2}, false); err != nil {
		t.Fatalf("expected dependent and prerequisite to reset together, got %v", err)
	}
}
func TestRemoveTaskDependencyUpdatesGraph(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := RemoveTaskDependency(1, 2); err != nil {
		t.Fatalf("remove dependency: %v", err)
	}
	if _, err := RemoveTaskDependency(1, 2); !errors.Is(err, ErrDependencyNotFound) {
		t.Fatalf("expected absent-edge error, got %v", err)
	}
	if _, err := RemoveTaskDependency(999, 2); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected missing-task error, got %v", err)
	}
}

func TestDeleteTaskRejectsReferencedPrerequisite(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	if DeleteTask(2) {
		t.Fatal("expected prerequisite deletion to be rejected")
	}
	if GetTaskByID(2) == nil {
		t.Fatal("referenced prerequisite must remain stored")
	}
	if _, ok := BulkDeleteTasks([]int{2}); ok {
		t.Fatal("expected bulk prerequisite deletion to be rejected")
	}
	if GetTaskByID(2) == nil {
		t.Fatal("rejected bulk delete must not mutate tasks")
	}
}

func TestDeleteTaskRecordsActivityBeforeRemoval(t *testing.T) {
	ResetTasks()
	if !DeleteTask(1) {
		t.Fatal("expected task deletion")
	}
	events := GetActivityLog()
	if len(events) != 1 || events[0].TaskID != 1 || events[0].Action != "deleted" || events[0].Summary != "Task deleted: Buy groceries" {
		t.Fatalf("unexpected deletion activity: %#v", events)
	}
}

func TestDeleteTaskRemovesTimeFromReports(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskComment(1, "Remove with task"); err != nil {
		t.Fatal(err)
	}
	startedAt := time.Now().Add(-2 * time.Hour)
	if _, err := AddManualTimeEntry(1, startedAt, startedAt.Add(time.Hour), "Completed work"); err != nil {
		t.Fatal(err)
	}
	if !DeleteTask(1) {
		t.Fatal("expected task deletion")
	}
	if report := GetTimeReport(nil, nil, nil); report.TotalSeconds != 0 || report.EntryCount != 0 || len(report.ByTask) != 0 {
		t.Fatalf("deleted task time remained in report: %#v", report)
	}
	if _, found := GetTaskComments(1, 0, 20); found {
		t.Fatal("expected comments to be removed with deleted task")
	}
}

func TestIsTaskBlockedFollowsPrerequisiteCompletion(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	blocked, found := IsTaskBlocked(1)
	if !found || !blocked {
		t.Fatalf("expected task to be blocked, got blocked=%v found=%v", blocked, found)
	}
	SetTaskCompletion(2, true)
	blocked, found = IsTaskBlocked(1)
	if !found || blocked {
		t.Fatalf("expected task to become unblocked, got blocked=%v found=%v", blocked, found)
	}
}

func TestSetTaskCompletionRejectsBlockedTask(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := SetTaskCompletionChecked(1, true); !errors.Is(err, ErrTaskBlocked) {
		t.Fatalf("expected blocked-task error, got %v", err)
	}
	if task := GetTaskByID(1); task == nil || task.Completed {
		t.Fatal("blocked completion must leave task incomplete")
	}
	if _, err := SetTaskCompletionChecked(2, true); err != nil {
		t.Fatalf("complete prerequisite: %v", err)
	}
	if _, err := SetTaskCompletionChecked(1, true); err != nil {
		t.Fatalf("complete unblocked task: %v", err)
	}
}

func TestBulkCompletionAcceptsPrerequisiteInSameBatch(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	result, err := BulkSetTaskCompletionChecked([]int{1, 2}, true)
	if err != nil || result.Updated != 2 {
		t.Fatalf("expected both tasks to complete atomically, result=%#v err=%v", result, err)
	}
}

func TestBulkCompletionCreatesNextOccurrenceAfterValidation(t *testing.T) {
	ResetTasks()
	recurring, err := CreateTaskWithDependencies(model.Task{
		Title: "Monthly close", DueDate: time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC),
		Recurrence: &model.RecurrenceRule{Frequency: "monthly", Interval: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BulkSetTaskCompletionChecked([]int{recurring.ID, 999}, true); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected invalid batch rejection, got %v", err)
	}
	if _, err := BulkSetTaskCompletionChecked([]int{recurring.ID}, true); err != nil {
		t.Fatal(err)
	}
	allTasks := GetAllTasks()
	occurrences := 0
	var next model.Task
	for _, task := range allTasks {
		if task.RecurrenceSeriesID == recurring.ID {
			occurrences++
			if task.RecurrenceOccurrence == 2 {
				next = task
			}
		}
	}
	if occurrences != 2 || next.DueDate.Format("2006-01-02") != "2026-11-30" {
		t.Fatalf("expected clamped next occurrence after successful batch, got %#v", allTasks)
	}
}

func TestRecurrenceUntilIncludesFinalDueDateAndThenStops(t *testing.T) {
	ResetTasks()
	firstDue := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	until := firstDue.AddDate(0, 0, 2)
	first, err := CreateTaskWithDependencies(model.Task{
		Title: "Short series", DueDate: firstDue,
		Recurrence: &model.RecurrenceRule{Frequency: "daily", Interval: 1, Until: &until},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SetTaskCompletionChecked(first.ID, true); err != nil {
		t.Fatal(err)
	}
	second := GetTaskByID(first.ID + 1)
	if second == nil || !second.DueDate.Equal(firstDue.AddDate(0, 0, 1)) {
		t.Fatalf("expected second occurrence, got %#v", second)
	}
	if _, err := SetTaskCompletionChecked(second.ID, true); err != nil {
		t.Fatal(err)
	}
	third := GetTaskByID(second.ID + 1)
	if third == nil || !third.DueDate.Equal(until) {
		t.Fatalf("expected final occurrence on until date, got %#v", third)
	}
	if _, err := SetTaskCompletionChecked(third.ID, true); err != nil {
		t.Fatal(err)
	}
	if GetTaskByID(third.ID+1) != nil {
		t.Fatal("expected recurrence to stop after inclusive until date")
	}
}

func TestGeneratedOccurrenceIsRecordedInActivity(t *testing.T) {
	ResetTasks()
	first, err := CreateTaskWithDependencies(model.Task{
		Title: "Daily cleanup", DueDate: time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC),
		Recurrence: &model.RecurrenceRule{Frequency: "daily", Interval: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SetTaskCompletionChecked(first.ID, true); err != nil {
		t.Fatal(err)
	}
	page := GetActivities(ptrInt(first.ID), "recurrence_created", nil, nil, 0, 10)
	if page.Total != 1 || !strings.Contains(page.Activities[0].Summary, strconv.Itoa(first.ID+1)) {
		t.Fatalf("expected recurrence creation activity, got %#v", page)
	}
}

func TestBlockedRecurringTaskDoesNotSpawnOccurrence(t *testing.T) {
	ResetTasks()
	first, err := CreateTaskWithDependencies(model.Task{
		Title: "Recurring close", DueDate: time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC),
		Recurrence: &model.RecurrenceRule{Frequency: "weekly", Interval: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AddTaskDependency(first.ID, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := SetTaskCompletionChecked(first.ID, true); !errors.Is(err, ErrTaskBlocked) {
		t.Fatalf("expected blocked completion, got %v", err)
	}
	if GetTaskByID(first.ID+1) != nil {
		t.Fatal("blocked completion must not spawn a recurrence")
	}
}

func TestBulkCompletionRejectsBlockedBatchWithoutMutation(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := BulkSetTaskCompletionChecked([]int{1}, true); !errors.Is(err, ErrTaskBlocked) {
		t.Fatalf("expected blocked-task error, got %v", err)
	}
	if task := GetTaskByID(1); task == nil || task.Completed {
		t.Fatal("rejected batch must leave task incomplete")
	}
}

func TestCompletionTransitionsRecordOnce(t *testing.T) {
	ResetTasks()
	if _, err := SetTaskCompletionChecked(1, true); err != nil {
		t.Fatal(err)
	}
	if _, err := SetTaskCompletionChecked(1, true); err != nil {
		t.Fatal(err)
	}
	if _, err := SetTaskCompletionChecked(1, false); err != nil {
		t.Fatal(err)
	}
	events := GetActivityLog()
	if len(events) != 2 || events[0].Action != "completed" || events[1].Action != "reopened" {
		t.Fatalf("expected one event per transition, got %#v", events)
	}
}

func TestGetReadyTasksExcludesBlockedAndCompletedTasks(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	ready := GetReadyTasks()
	if len(ready) != 1 || ready[0].ID != 2 {
		t.Fatalf("expected only prerequisite task 2 to be ready, got %#v", ready)
	}
	if _, err := SetTaskCompletionChecked(2, true); err != nil {
		t.Fatal(err)
	}
	ready = GetReadyTasks()
	if len(ready) != 1 || ready[0].ID != 1 {
		t.Fatalf("expected completed task excluded and dependent ready, got %#v", ready)
	}
}

func TestGetTasksReturnsEmptyPageBeyondResults(t *testing.T) {
	page := GetTasks(nil, nil, nil, nil, nil, "", nil, nil, nil, nil, nil, nil, 100, 20, "id", false)
	if page == nil {
		t.Fatal("expected an empty slice, got nil")
	}
	if len(page) != 0 {
		t.Fatalf("expected empty page, got %d tasks", len(page))
	}
}

func TestPaginatedQueriesHandleMaximumOffset(t *testing.T) {
	ResetTasks()
	maxInt := int(^uint(0) >> 1)
	if page := GetTasks(nil, nil, nil, nil, nil, "", nil, nil, nil, nil, nil, nil, maxInt, 20, "id", false); len(page) != 0 {
		t.Fatalf("expected empty task page for huge offset, got %#v", page)
	}
	if page, found := GetTaskComments(1, maxInt, 20); !found || len(page.Comments) != 0 {
		t.Fatalf("expected empty comment page for huge offset, page=%#v found=%v", page, found)
	}
	if page := GetTagSummary(maxInt, 20); len(page.Tags) != 0 {
		t.Fatalf("expected empty tag page for huge offset, got %#v", page)
	}
	if page := GetActivities(nil, "", nil, nil, maxInt, 20); len(page.Activities) != 0 {
		t.Fatalf("expected empty activity page for huge offset, got %#v", page)
	}
}

func TestGetUpcomingTasksFiltersAndSorts(t *testing.T) {
	ResetTasks()
	now := time.Date(2030, 1, 1, 12, 0, 0, 0, time.UTC)
	first, err := CreateTaskWithDependencies(model.Task{Title: "Soonest", DueDate: now.Add(24 * time.Hour)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := CreateTaskWithDependencies(model.Task{Title: "Later", DueDate: now.Add(48 * time.Hour)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	completed, err := CreateTaskWithDependencies(model.Task{Title: "Done", DueDate: now.Add(12 * time.Hour)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	SetTaskCompletion(completed.ID, true)
	result := GetUpcomingTasks(now, 3)
	if len(result) != 2 || result[0].ID != first.ID || result[1].ID != second.ID {
		t.Fatalf("unexpected upcoming tasks: %#v", result)
	}
}

func TestGetBlockedTasksReturnsIncompleteDependents(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	blocked := GetBlockedTasks()
	if len(blocked) != 1 || blocked[0].ID != 1 || blocked[0].Completed {
		t.Fatalf("unexpected blocked task list: %#v", blocked)
	}
	if _, err := SetTaskCompletionChecked(2, true); err != nil {
		t.Fatal(err)
	}
	if blocked := GetBlockedTasks(); len(blocked) != 0 {
		t.Fatalf("expected no blocked tasks after prerequisite completion, got %#v", blocked)
	}
}

func TestCanceledTasksStayOutOfActiveQueuesAndPendingCounts(t *testing.T) {
	ResetTasks()
	dueDate := time.Date(2030, 1, 1, 9, 0, 0, 0, time.UTC)
	canceled, err := CreateTaskWithDependencies(model.Task{Title: "Canceled", DueDate: dueDate, Status: model.TaskStatusCanceled}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CreateTaskWithDependencies(model.Task{Title: "Active", DueDate: dueDate.Add(time.Hour)}, nil); err != nil {
		t.Fatal(err)
	}
	for _, task := range GetReadyTasks() {
		if task.ID == canceled.ID {
			t.Fatal("canceled task appeared in ready queue")
		}
	}
	for _, task := range GetUpcomingTasks(time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC), 1) {
		if task.ID == canceled.ID {
			t.Fatal("canceled task appeared in upcoming queue")
		}
	}
	summary := GetTaskSummary()
	if summary.Pending != 3 || summary.ByStatus[model.TaskStatusCanceled] != 1 {
		t.Fatalf("canceled task counted as pending: %#v", summary)
	}
}

func TestCanceledTasksAreNotReportedAsBlocked(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	updated, err := UpdateTaskWithDependencies(1, model.Task{Title: "Canceled dependent", DueDate: time.Now(), Status: model.TaskStatusCanceled, DependsOn: []int{2}}, []int{2})
	if err != nil || updated == nil {
		t.Fatalf("expected cancellation update, task=%#v err=%v", updated, err)
	}
	if blocked := GetBlockedTasks(); len(blocked) != 0 {
		t.Fatalf("canceled task appeared in blocked queue: %#v", blocked)
	}
	if summary := GetTaskSummary(); summary.Blocked != 0 {
		t.Fatalf("canceled task counted as blocked: %#v", summary)
	}
}

func TestBulkSetTaskCompletionUpdatesAllTasks(t *testing.T) {
	ResetTasks()

	result, ok := BulkSetTaskCompletion([]int{1, 2}, true)
	if !ok {
		t.Fatal("expected bulk update to succeed")
	}
	if result.Updated != 2 || len(result.Tasks) != 2 {
		t.Fatalf("expected two updated tasks, got %#v", result)
	}
	for _, task := range result.Tasks {
		if !task.Completed {
			t.Fatalf("expected task %d to be completed", task.ID)
		}
	}
}

func TestBulkSetTaskCompletionDoesNotPartiallyUpdate(t *testing.T) {
	ResetTasks()

	if _, ok := BulkSetTaskCompletion([]int{1, 999}, true); ok {
		t.Fatal("expected bulk update to reject missing task")
	}
	if task := GetTaskByID(1); task == nil || task.Completed {
		t.Fatal("expected task 1 to remain unchanged")
	}
}

func TestBulkSetTaskStatusIsAtomicAndDependencyAware(t *testing.T) {
	ResetTasks()
	if _, err := AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := BulkSetTaskStatusChecked([]int{1}, model.TaskStatusCompleted); !errors.Is(err, ErrTaskBlocked) {
		t.Fatalf("expected blocked status update, got %v", err)
	}
	if task := GetTaskByID(1); task == nil || task.Status != model.TaskStatusTodo {
		t.Fatalf("blocked bulk update mutated task: %#v", task)
	}
	result, err := BulkSetTaskStatusChecked([]int{1, 2}, model.TaskStatusInProgress)
	if err != nil || result.Updated != 2 || result.Tasks[0].Status != model.TaskStatusInProgress {
		t.Fatalf("unexpected bulk status result: %#v err=%v", result, err)
	}
	if _, err := BulkSetTaskStatusChecked([]int{1}, "blocked"); !errors.Is(err, ErrTaskStatusInvalid) {
		t.Fatalf("expected invalid status error, got %v", err)
	}
}

func TestBulkDeleteTasksIsAtomic(t *testing.T) {
	ResetTasks()
	if _, ok := BulkDeleteTasks([]int{1, 999}); ok {
		t.Fatal("expected deletion to reject a missing task")
	}
	if GetTaskByID(1) == nil {
		t.Fatal("expected existing task to remain after rejected deletion")
	}
	result, ok := BulkDeleteTasks([]int{1, 2})
	if !ok || result.Count != 2 || len(result.Deleted) != 2 {
		t.Fatalf("expected both tasks deleted, got %#v, ok=%v", result, ok)
	}
	if len(GetAllTasks()) != 0 {
		t.Fatal("expected no tasks to remain")
	}
}

func TestBulkDeleteTasksRemovesRelatedTimeAndComments(t *testing.T) {
	ResetTasks()
	start := time.Now().Add(-2 * time.Hour)
	if _, err := AddManualTimeEntry(1, start, start.Add(time.Hour), "Tracked"); err != nil {
		t.Fatal(err)
	}
	if _, err := AddTaskComment(1, "Note"); err != nil {
		t.Fatal(err)
	}
	if _, ok := BulkDeleteTasks([]int{1}); !ok {
		t.Fatal("expected task deletion")
	}
	if report := GetTimeReport(nil, nil, nil); report.TotalSeconds != 0 || report.EntryCount != 0 {
		t.Fatalf("bulk-deleted task time remained in report: %#v", report)
	}
	if _, found := GetTaskComments(1, 0, 20); found {
		t.Fatal("expected comments to be removed with their task")
	}
}

func TestCreateTaskNormalizesTags(t *testing.T) {
	ResetTasks()
	created := CreateTask(model.Task{Title: "Tagged", Tags: []string{" home ", "Home", "work"}})
	if len(created.Tags) != 2 || created.Tags[0] != "home" || created.Tags[1] != "work" {
		t.Fatalf("expected trimmed, unique tags, got %#v", created.Tags)
	}
}

func TestDuplicateTaskResetsProgressAndDetachesRecurrence(t *testing.T) {
	ResetTasks()
	source, err := CreateTaskWithDependencies(model.Task{
		Title: "Weekly report", DueDate: time.Now().Add(24 * time.Hour), Completed: true,
		Recurrence: &model.RecurrenceRule{Frequency: "weekly", Interval: 1},
		Checklist:  []model.ChecklistItem{{ID: 90, Text: "Draft", Completed: true}},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	duplicate, err := DuplicateTask(source.ID)
	if err != nil {
		t.Fatal(err)
	}
	if duplicate.ID == source.ID || duplicate.Title != "Weekly report (copy)" || duplicate.Completed || duplicate.Status != model.TaskStatusTodo || duplicate.Recurrence != nil {
		t.Fatalf("unexpected duplicate task state: %#v", duplicate)
	}
	if len(duplicate.Checklist) != 1 || duplicate.Checklist[0].ID == source.Checklist[0].ID || duplicate.Checklist[0].Completed {
		t.Fatalf("duplicate checklist progress/identity was not reset: %#v", duplicate.Checklist)
	}
	if !IsActivityAction("duplicated") {
		t.Fatal("duplicate activity must be supported")
	}
}

func TestCreateTaskDefaultsPriority(t *testing.T) {
	ResetTasks()
	created := CreateTask(model.Task{Title: "No priority"})
	if created.Priority != "medium" {
		t.Fatalf("expected medium priority by default, got %q", created.Priority)
	}
}

func TestCreateTaskStartsRecurrenceSeries(t *testing.T) {
	ResetTasks()
	task, err := CreateTaskWithDependencies(model.Task{
		Title:      "Weekly review",
		DueDate:    time.Date(2026, 10, 2, 9, 0, 0, 0, time.UTC),
		Recurrence: &model.RecurrenceRule{Frequency: "weekly", Interval: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if task.RecurrenceSeriesID != task.ID || task.RecurrenceOccurrence != 1 {
		t.Fatalf("expected service-owned first occurrence metadata, got %#v", task)
	}
}

func TestCompletingRecurringTaskCreatesExactlyOneNextOccurrence(t *testing.T) {
	ResetTasks()
	first, err := CreateTaskWithDependencies(model.Task{
		Title: "Weekly review", DueDate: time.Date(2026, 10, 2, 9, 0, 0, 0, time.UTC),
		Recurrence: &model.RecurrenceRule{Frequency: "weekly", Interval: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SetTaskCompletionChecked(first.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := SetTaskCompletionChecked(first.ID, true); err != nil {
		t.Fatal(err)
	}
	allTasks := GetAllTasks()
	occurrenceCount := 0
	var next model.Task
	for _, task := range allTasks {
		if task.RecurrenceSeriesID == first.ID {
			occurrenceCount++
			if task.RecurrenceOccurrence == 2 {
				next = task
			}
		}
	}
	if occurrenceCount != 2 || next.RecurrenceOccurrence != 2 || !next.DueDate.Equal(first.DueDate.AddDate(0, 0, 7)) {
		t.Fatalf("expected one next weekly occurrence, got %#v", allTasks)
	}
}

func TestGetTaskOccurrencesReturnsOrderedSeriesCopies(t *testing.T) {
	ResetTasks()
	first, err := CreateTaskWithDependencies(model.Task{
		Title: "Daily report", DueDate: time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC),
		Recurrence: &model.RecurrenceRule{Frequency: "daily", Interval: 1},
		Tags:       []string{"work"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SetTaskCompletionChecked(first.ID, true); err != nil {
		t.Fatal(err)
	}
	occurrences, err := GetTaskOccurrences(first.ID + 1)
	if err != nil || len(occurrences) != 2 || occurrences[0].RecurrenceOccurrence != 1 || occurrences[1].RecurrenceOccurrence != 2 {
		t.Fatalf("unexpected occurrence list: %#v err=%v", occurrences, err)
	}
	occurrences[0].Tags[0] = "mutated"
	stored := GetTaskByID(first.ID)
	if stored.Tags[0] != "work" {
		t.Fatal("occurrence query exposed mutable internal slices")
	}
}

func TestGetNextOccurrenceDueDateDoesNotMutateSeries(t *testing.T) {
	ResetTasks()
	until := time.Date(2026, 10, 3, 9, 0, 0, 0, time.UTC)
	task, err := CreateTaskWithDependencies(model.Task{
		Title: "Daily reminder", DueDate: time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC),
		Recurrence: &model.RecurrenceRule{Frequency: "daily", Interval: 1, Until: &until},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	next, err := GetNextOccurrenceDueDate(task.ID)
	if err != nil || !next.Equal(task.DueDate.AddDate(0, 0, 1)) {
		t.Fatalf("unexpected next due date %v err=%v", next, err)
	}
	if len(GetAllTasks()) != 3 {
		t.Fatal("preview must not create an occurrence")
	}
	if _, err := GetNextOccurrenceDueDate(2); !errors.Is(err, ErrRecurrenceNotFound) {
		t.Fatalf("expected non-recurring-task error, got %v", err)
	}
}

func TestBuildNextOccurrenceResetsMutableProgress(t *testing.T) {
	due := time.Date(2026, 10, 9, 9, 0, 0, 0, time.UTC)
	source := model.Task{
		ID: 12, Title: "Weekly review", DueDate: due.AddDate(0, 0, -7), Completed: true,
		Tags: []string{"team"}, DependsOn: []int{2}, Recurrence: &model.RecurrenceRule{Frequency: "weekly", Interval: 1},
		RecurrenceSeriesID: 12, RecurrenceOccurrence: 1,
		Checklist: []model.ChecklistItem{{ID: 4, Text: "Prepare", Completed: true}, {ID: 5, Text: "Review", Completed: false}},
	}
	now := time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC)
	next, nextChecklistID := buildNextOccurrence(source, 13, 40, due, now)
	if next.ID != 13 || next.DueDate != due || next.Completed || next.RecurrenceOccurrence != 2 || next.RecurrenceSeriesID != 12 {
		t.Fatalf("unexpected next occurrence metadata: %#v", next)
	}
	if len(next.Checklist) != 2 || next.Checklist[0].ID != 40 || next.Checklist[0].Completed || next.Checklist[1].ID != 41 || nextChecklistID != 42 {
		t.Fatalf("unexpected cloned checklist: %#v next ID %d", next.Checklist, nextChecklistID)
	}
}

func TestCreateTaskRecordsActivity(t *testing.T) {
	ResetTasks()
	created, err := CreateTaskWithDependencies(model.Task{Title: "Plan release"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	events := GetActivityLog()
	if len(events) != 1 || events[0].TaskID != created.ID || events[0].Action != "created" || events[0].Summary != "Task created: Plan release" {
		t.Fatalf("unexpected creation activity: %#v", events)
	}
}

func TestBulkSetTaskPriorityIsAtomic(t *testing.T) {
	ResetTasks()
	if _, ok := BulkSetTaskPriority([]int{1, 999}, "high"); ok {
		t.Fatal("expected update to reject a missing task")
	}
	if task := GetTaskByID(1); task == nil || task.Priority != "medium" {
		t.Fatal("expected existing task priority to remain unchanged")
	}
	result, ok := BulkSetTaskPriority([]int{1, 2}, "high")
	if !ok || result.Updated != 2 {
		t.Fatalf("expected two tasks updated, got %#v, ok=%v", result, ok)
	}
}

func TestBulkSetTaskPriorityRecordsOnlyChanges(t *testing.T) {
	ResetTasks()
	if _, ok := BulkSetTaskPriority([]int{1, 2}, "high"); !ok {
		t.Fatal("expected priority update")
	}
	if events := GetActivities(nil, "updated", nil, nil, 0, 20); events.Total != 2 {
		t.Fatalf("expected one activity per changed task, got %#v", events)
	}
	if _, ok := BulkSetTaskPriority([]int{1, 2}, "high"); !ok {
		t.Fatal("expected repeated priority update")
	}
	if events := GetActivities(nil, "updated", nil, nil, 0, 20); events.Total != 2 {
		t.Fatalf("no-op update should not add activity, got %#v", events)
	}
}

func TestBulkSetTaskDueDateIsAtomic(t *testing.T) {
	ResetTasks()
	dueDate := time.Date(2030, 1, 15, 9, 0, 0, 0, time.UTC)
	if _, ok := BulkSetTaskDueDate([]int{1, 999}, dueDate); ok {
		t.Fatal("expected update to reject a missing task")
	}
	if task := GetTaskByID(1); task == nil || task.DueDate.Equal(dueDate) {
		t.Fatal("rejected update must not partially change due dates")
	}
	result, ok := BulkSetTaskDueDate([]int{1, 2, 1}, dueDate)
	if !ok || result.Updated != 2 {
		t.Fatalf("expected two unique tasks updated, got %#v, ok=%v", result, ok)
	}
	for _, task := range result.Tasks {
		if !task.DueDate.Equal(dueDate) {
			t.Fatalf("expected due date %v, got %v", dueDate, task.DueDate)
		}
	}
}

func TestBulkSetTaskDueDateRecordsChanges(t *testing.T) {
	ResetTasks()
	dueDate := time.Date(2030, 1, 15, 9, 0, 0, 0, time.UTC)
	if _, ok := BulkSetTaskDueDate([]int{1, 2}, dueDate); !ok {
		t.Fatal("expected due-date update")
	}
	if events := GetActivities(nil, "updated", nil, nil, 0, 20); events.Total != 2 {
		t.Fatalf("expected due-date activity for both tasks, got %#v", events)
	}
	if _, ok := BulkSetTaskDueDate([]int{1, 2}, dueDate); !ok {
		t.Fatal("expected repeated due-date update")
	}
	if events := GetActivities(nil, "updated", nil, nil, 0, 20); events.Total != 2 {
		t.Fatalf("no-op update should not add activity, got %#v", events)
	}
}

func TestBulkSetTaskTagsIsAtomicAndNormalizes(t *testing.T) {
	ResetTasks()
	if _, ok := BulkSetTaskTags([]int{1, 999}, []string{" work ", "WORK"}); ok {
		t.Fatal("expected update to reject a missing task")
	}
	if task := GetTaskByID(1); task == nil || task.Tags[0] != "home" {
		t.Fatal("rejected update must not change tags")
	}
	result, ok := BulkSetTaskTags([]int{1, 2}, []string{" work ", "WORK", "personal"})
	if !ok || result.Updated != 2 {
		t.Fatalf("expected two tasks updated, got %#v, ok=%v", result, ok)
	}
	for _, task := range result.Tasks {
		if len(task.Tags) != 2 || task.Tags[0] != "work" || task.Tags[1] != "personal" {
			t.Fatalf("unexpected normalized tags: %#v", task.Tags)
		}
	}
}

func TestBulkSetTaskTagsRecordsOnlyChangedSets(t *testing.T) {
	ResetTasks()
	if _, ok := BulkSetTaskTags([]int{1}, []string{"HOME", "errands"}); !ok {
		t.Fatal("expected tag update")
	}
	if events := GetActivities(nil, "updated", nil, nil, 0, 20); events.Total != 0 {
		t.Fatalf("case/order-equivalent tags should not add activity: %#v", events)
	}
	if _, ok := BulkSetTaskTags([]int{1}, []string{"work"}); !ok {
		t.Fatal("expected tag replacement")
	}
	if events := GetActivities(nil, "updated", nil, nil, 0, 20); events.Total != 1 {
		t.Fatalf("expected one tag-change event, got %#v", events)
	}
}

func TestBulkSetTaskEstimateIsAtomic(t *testing.T) {
	ResetTasks()
	if _, err := BulkSetTaskEstimate([]int{1, 999}, 45); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected missing-task error, got %v", err)
	}
	if task := GetTaskByID(1); task == nil || task.EstimateMinutes != 0 {
		t.Fatalf("rejected batch changed estimate: %#v", task)
	}
	result, err := BulkSetTaskEstimate([]int{1, 2, 1}, 45)
	if err != nil || result.Updated != 2 || result.Tasks[0].EstimateMinutes != 45 {
		t.Fatalf("unexpected bulk estimate result: %#v err=%v", result, err)
	}
	if _, err := BulkSetTaskEstimate([]int{1}, -1); !errors.Is(err, ErrTaskEstimateInvalid) {
		t.Fatalf("expected negative estimate rejection, got %v", err)
	}
}

func TestUpdateTaskPreservesCreatedAt(t *testing.T) {
	ResetTasks()
	original := GetTaskByID(1)
	if original.CreatedAt.IsZero() || original.UpdatedAt.IsZero() {
		t.Fatal("expected seeded task timestamps to be initialized")
	}
	updated := UpdateTask(1, model.Task{Title: "Updated", DueDate: original.DueDate})
	if updated == nil {
		t.Fatal("expected task to update")
	}
	if !updated.CreatedAt.Equal(original.CreatedAt) {
		t.Fatalf("expected createdAt %v to be preserved, got %v", original.CreatedAt, updated.CreatedAt)
	}
	if updated.UpdatedAt.IsZero() {
		t.Fatal("expected updatedAt to be set")
	}
}

func TestUpdateTaskCompletionRecordsLifecycleAndRecurrence(t *testing.T) {
	ResetTasks()
	dueDate := time.Date(2030, 1, 1, 9, 0, 0, 0, time.UTC)
	task, err := CreateTaskWithDependencies(model.Task{
		Title: "Daily review", DueDate: dueDate,
		Recurrence: &model.RecurrenceRule{Frequency: "daily", Interval: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	task.Completed = true
	updated, err := UpdateTaskWithDependencies(task.ID, task, task.DependsOn)
	if err != nil || updated == nil || !updated.Completed {
		t.Fatalf("expected task completion update, task=%#v err=%v", updated, err)
	}
	occurrences, err := GetTaskOccurrences(task.ID)
	if err != nil || len(occurrences) != 2 || !occurrences[1].DueDate.Equal(dueDate.AddDate(0, 0, 1)) {
		t.Fatalf("expected next recurrence, got %#v err=%v", occurrences, err)
	}
	events := GetActivities(&task.ID, "completed", nil, nil, 0, 20)
	if events.Total != 1 {
		t.Fatalf("expected one completion event, got %#v", events)
	}
}

func TestSkipTaskOccurrenceCancelsAndAdvancesSeries(t *testing.T) {
	ResetTasks()
	dueDate := time.Date(2030, 1, 1, 9, 0, 0, 0, time.UTC)
	task, err := CreateTaskWithDependencies(model.Task{Title: "Weekly review", DueDate: dueDate, Recurrence: &model.RecurrenceRule{Frequency: "weekly", Interval: 1}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	skipped, err := SkipTaskOccurrence(task.ID)
	if err != nil || skipped.Status != model.TaskStatusCanceled || skipped.Completed {
		t.Fatalf("unexpected skipped occurrence: %#v err=%v", skipped, err)
	}
	occurrences, err := GetTaskOccurrences(task.ID)
	if err != nil || len(occurrences) != 2 || occurrences[1].DueDate != dueDate.AddDate(0, 0, 7) || occurrences[1].Status != model.TaskStatusTodo {
		t.Fatalf("expected next weekly occurrence, got %#v err=%v", occurrences, err)
	}
	if events := GetActivities(&task.ID, "recurrence_skipped", nil, nil, 0, 20); events.Total != 1 {
		t.Fatalf("expected skip activity, got %#v", events)
	}
	if _, err := SkipTaskOccurrence(task.ID); !errors.Is(err, ErrTaskAlreadyClosed) {
		t.Fatalf("expected repeated skip to fail, got %v", err)
	}
}

func TestUpdateTaskEstimateValidatesAndRecordsChange(t *testing.T) {
	ResetTasks()
	updated, err := UpdateTaskEstimate(1, 45)
	if err != nil || updated.EstimateMinutes != 45 {
		t.Fatalf("expected estimate update, task=%#v err=%v", updated, err)
	}
	if _, err := UpdateTaskEstimate(1, -1); !errors.Is(err, ErrTaskEstimateInvalid) {
		t.Fatalf("expected negative estimate error, got %v", err)
	}
	events := GetActivityLog()
	if len(events) != 1 || events[0].Action != "estimate_updated" {
		t.Fatalf("expected one estimate activity event, got %#v", events)
	}
}

func TestGetTaskTimeSummaryIncludesTrackedAndRunningTime(t *testing.T) {
	ResetTasks()
	if _, err := UpdateTaskEstimate(1, 60); err != nil {
		t.Fatal(err)
	}
	start := time.Now().UTC().Add(-30 * time.Minute)
	if _, err := AddManualTimeEntry(1, start, start.Add(15*time.Minute), "Earlier"); err != nil {
		t.Fatal(err)
	}
	if _, err := StartTaskTimer(1, "Current"); err != nil {
		t.Fatal(err)
	}
	summary, found := GetTaskTimeSummary(1)
	if !found || summary.EstimatedSeconds != 3600 || summary.ActualSeconds < 900 || summary.VarianceSeconds != summary.ActualSeconds-3600 {
		t.Fatalf("unexpected task time summary: %#v found=%v", summary, found)
	}
	if _, found := GetTaskTimeSummary(999); found {
		t.Fatal("expected missing task summary to be reported")
	}
}

func TestGetTimeReportClipsIntervalsToWindow(t *testing.T) {
	ResetTasks()
	start := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	if _, err := AddManualTimeEntry(1, start, end, "Planning"); err != nil {
		t.Fatal(err)
	}
	from, to := start.Add(15*time.Minute), start.Add(45*time.Minute)
	report := GetTimeReport(nil, &from, &to)
	if report.TotalSeconds != 1800 || report.EntryCount != 1 || len(report.ByTask) != 1 || report.ByTask[0].TaskID != 1 {
		t.Fatalf("unexpected clipped time report: %#v", report)
	}
}

func TestTimeReportExcludesActiveTimerOutsideWindow(t *testing.T) {
	ResetTasks()
	if _, err := StartTaskTimer(1, "Current"); err != nil {
		t.Fatal(err)
	}
	from := time.Now().UTC().Add(time.Hour)
	report := GetTimeReport(nil, &from, nil)
	if report.ActiveCount != 0 || report.TotalSeconds != 0 {
		t.Fatalf("future window must exclude current timer, got %#v", report)
	}
}

func TestUpdateTaskRecordsRenameActivity(t *testing.T) {
	ResetTasks()
	original := GetTaskByID(1)
	updated := UpdateTask(1, model.Task{Title: "Market run", DueDate: original.DueDate, DependsOn: original.DependsOn})
	if updated == nil {
		t.Fatal("expected task to update")
	}
	events := GetActivityLog()
	if len(events) != 1 || events[0].Action != "updated" || events[0].Summary != "Task renamed from Buy groceries to Market run" {
		t.Fatalf("unexpected update activity: %#v", events)
	}
}
