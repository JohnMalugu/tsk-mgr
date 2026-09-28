package service

import (
	"errors"
	"testing"

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

func TestChecklistErrorsAreDistinct(t *testing.T) {
	if errors.Is(ErrChecklistItemNotFound, ErrTaskNotFound) {
		t.Fatal("missing checklist item and missing task must be distinct")
	}
	if ErrChecklistOrderInvalid.Error() != "checklist order must contain every item exactly once" {
		t.Fatalf("unexpected checklist order error: %q", ErrChecklistOrderInvalid)
	}
}

func TestGetTaskByIDReturnsCopy(t *testing.T) {
	task := GetTaskByID(1)
	if task == nil {
		t.Fatal("expected seeded task")
	}

	originalTitle := task.Title
	task.Title = "locally changed"

	storedTask := GetTaskByID(1)
	if storedTask == nil {
		t.Fatal("expected seeded task")
	}
	if storedTask.Title != originalTitle {
		t.Fatalf("expected stored task title %q, got %q", originalTitle, storedTask.Title)
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
	page := GetTasks(nil, nil, "", nil, nil, nil, nil, 100, 20, "id", false)
	if page == nil {
		t.Fatal("expected an empty slice, got nil")
	}
	if len(page) != 0 {
		t.Fatalf("expected empty page, got %d tasks", len(page))
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

func TestCreateTaskNormalizesTags(t *testing.T) {
	ResetTasks()
	created := CreateTask(model.Task{Title: "Tagged", Tags: []string{" home ", "Home", "work"}})
	if len(created.Tags) != 2 || created.Tags[0] != "home" || created.Tags[1] != "work" {
		t.Fatalf("expected trimmed, unique tags, got %#v", created.Tags)
	}
}

func TestCreateTaskDefaultsPriority(t *testing.T) {
	ResetTasks()
	created := CreateTask(model.Task{Title: "No priority"})
	if created.Priority != "medium" {
		t.Fatalf("expected medium priority by default, got %q", created.Priority)
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
