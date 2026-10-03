package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

func TestHealthRoute(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	Router(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected health response: %q", recorder.Body.String())
	}
}

func TestTimeTrackingReadRoutes(t *testing.T) {
	service.ResetTasks()
	entries := httptest.NewRecorder()
	Router(entries, httptest.NewRequest(http.MethodGet, "/tasks/1/time", nil))
	if entries.Code != http.StatusOK || !strings.Contains(entries.Body.String(), `"entries":[]`) {
		t.Fatalf("unexpected task time response: %d %s", entries.Code, entries.Body.String())
	}
	active := httptest.NewRecorder()
	Router(active, httptest.NewRequest(http.MethodGet, "/timer", nil))
	if active.Code != http.StatusOK || active.Body.String() != "null\n" {
		t.Fatalf("unexpected inactive timer response: %d %s", active.Code, active.Body.String())
	}
}

func TestTaskTimerStartAndStopRoutes(t *testing.T) {
	service.ResetTasks()
	started := httptest.NewRecorder()
	Router(started, httptest.NewRequest(http.MethodPost, "/tasks/1/timer/start", strings.NewReader(`{"note":"Focus"}`)))
	if started.Code != http.StatusCreated {
		t.Fatalf("timer start status %d: %s", started.Code, started.Body.String())
	}
	conflict := httptest.NewRecorder()
	Router(conflict, httptest.NewRequest(http.MethodPost, "/tasks/2/timer/start", nil))
	if conflict.Code != http.StatusConflict {
		t.Fatalf("expected active timer conflict %d, got %d", http.StatusConflict, conflict.Code)
	}
	stopped := httptest.NewRecorder()
	Router(stopped, httptest.NewRequest(http.MethodPost, "/tasks/1/timer/stop", nil))
	if stopped.Code != http.StatusOK || !strings.Contains(stopped.Body.String(), `"endedAt"`) {
		t.Fatalf("unexpected timer stop response: %d %s", stopped.Code, stopped.Body.String())
	}
}

func TestBulkDueDateRoute(t *testing.T) {
	service.ResetTasks()
	payload := `{"ids":[1,2],"dueDate":"2030-01-15T09:00:00Z"}`
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodPost, "/tasks/bulk/due-date", strings.NewReader(payload)))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"updated":2`) {
		t.Fatalf("unexpected bulk due-date response: %d %s", recorder.Code, recorder.Body.String())
	}
	invalid := httptest.NewRecorder()
	Router(invalid, httptest.NewRequest(http.MethodPost, "/tasks/bulk/due-date", strings.NewReader(`{"ids":[1],"dueDate":""}`)))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected missing date to be rejected, got %d", invalid.Code)
	}
}

func TestUpcomingTasksRouteValidatesWindow(t *testing.T) {
	service.ResetTasks()
	valid := httptest.NewRecorder()
	Router(valid, httptest.NewRequest(http.MethodGet, "/tasks/upcoming?days=7", nil))
	if valid.Code != http.StatusOK {
		t.Fatalf("expected upcoming task response, got %d %s", valid.Code, valid.Body.String())
	}
	invalid := httptest.NewRecorder()
	Router(invalid, httptest.NewRequest(http.MethodGet, "/tasks/upcoming?days=0", nil))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid window rejection, got %d", invalid.Code)
	}
}

func TestBulkTagsRoute(t *testing.T) {
	service.ResetTasks()
	payload := `{"ids":[1,2],"tags":[" work ","WORK","personal"]}`
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodPost, "/tasks/bulk/tags", strings.NewReader(payload)))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"updated":2`) || strings.Count(recorder.Body.String(), `"work"`) != 2 {
		t.Fatalf("unexpected bulk tags response: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestBulkStatusRoute(t *testing.T) {
	service.ResetTasks()
	payload := `{"ids":[1,2],"status":"in_progress"}`
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodPost, "/tasks/bulk/status", strings.NewReader(payload)))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"updated":2`) {
		t.Fatalf("unexpected bulk status response: %d %s", recorder.Code, recorder.Body.String())
	}
	invalid := httptest.NewRecorder()
	Router(invalid, httptest.NewRequest(http.MethodPost, "/tasks/bulk/status", strings.NewReader(`{"ids":[1],"status":"blocked"}`)))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid status rejection, got %d", invalid.Code)
	}
}

func TestManualTimeEntryRouteAndDeletion(t *testing.T) {
	service.ResetTasks()
	body := `{"startedAt":"2026-09-30T09:00:00Z","endedAt":"2026-09-30T09:30:00Z","note":"Planning"}`
	created := httptest.NewRecorder()
	Router(created, httptest.NewRequest(http.MethodPost, "/tasks/1/time", strings.NewReader(body)))
	if created.Code != http.StatusCreated {
		t.Fatalf("manual entry create status %d: %s", created.Code, created.Body.String())
	}
	var entry model.TimeEntry
	if err := json.NewDecoder(created.Body).Decode(&entry); err != nil {
		t.Fatal(err)
	}
	deleted := httptest.NewRecorder()
	Router(deleted, httptest.NewRequest(http.MethodDelete, "/tasks/1/time/"+fmt.Sprint(entry.ID), nil))
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("expected deletion status %d, got %d", http.StatusNoContent, deleted.Code)
	}
}

func TestTimeReportRouteFiltersByTaskAndTime(t *testing.T) {
	service.ResetTasks()
	start := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	service.AddManualTimeEntry(1, start, start.Add(time.Hour), "Planning")
	url := "/time/report?taskId=1&from=2026-09-30T09:15:00Z&to=2026-09-30T09:45:00Z"
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodGet, url, nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"totalSeconds":1800`) {
		t.Fatalf("unexpected time report response: %d %s", recorder.Code, recorder.Body.String())
	}
	invalid := httptest.NewRecorder()
	Router(invalid, httptest.NewRequest(http.MethodGet, "/time/report?taskId=0", nil))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid task filter status %d, got %d", http.StatusBadRequest, invalid.Code)
	}
}

func TestGlobalActivityRoute(t *testing.T) {
	service.ResetTasks()
	service.CreateTask(model.Task{Title: "Timeline entry"})
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodGet, "/activity", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"action":"created"`) {
		t.Fatalf("unexpected activity route response: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestTaskActivityRouteIncludesTaskEvents(t *testing.T) {
	service.ResetTasks()
	created := service.CreateTask(model.Task{Title: "Timeline entry"})
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodGet, "/tasks/"+fmt.Sprint(created.ID)+"/activity", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "Task created: Timeline entry") {
		t.Fatalf("unexpected task activity response: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestTaskCommentsRouteLifecycle(t *testing.T) {
	service.ResetTasks()
	created := httptest.NewRecorder()
	Router(created, httptest.NewRequest(http.MethodPost, "/tasks/1/comments", strings.NewReader(`{"body":"  Keep the scope focused  "}`)))
	if created.Code != http.StatusCreated || !strings.Contains(created.Body.String(), `"body":"Keep the scope focused"`) {
		t.Fatalf("unexpected comment creation response: %d %s", created.Code, created.Body.String())
	}
	listed := httptest.NewRecorder()
	Router(listed, httptest.NewRequest(http.MethodGet, "/tasks/1/comments", nil))
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), "Keep the scope focused") {
		t.Fatalf("unexpected comments list response: %d %s", listed.Code, listed.Body.String())
	}
	invalid := httptest.NewRecorder()
	Router(invalid, httptest.NewRequest(http.MethodPost, "/tasks/1/comments", strings.NewReader(`{"body":" "}`)))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected empty comment rejection, got %d", invalid.Code)
	}
}

func TestTaskOccurrencesRouteReturnsSeries(t *testing.T) {
	service.ResetTasks()
	first, err := service.CreateTaskWithDependencies(model.Task{
		Title: "Daily report", DueDate: time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC),
		Recurrence: &model.RecurrenceRule{Frequency: "daily", Interval: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetTaskCompletionChecked(first.ID, true); err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodGet, "/tasks/"+fmt.Sprint(first.ID)+"/occurrences", nil))
	if recorder.Code != http.StatusOK || strings.Count(recorder.Body.String(), `"recurrenceOccurrence"`) != 2 {
		t.Fatalf("unexpected occurrences response: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestNextRecurrenceRoutePreviewsDate(t *testing.T) {
	service.ResetTasks()
	task, err := service.CreateTaskWithDependencies(model.Task{
		Title: "Weekly report", DueDate: time.Date(2026, 10, 2, 9, 0, 0, 0, time.UTC),
		Recurrence: &model.RecurrenceRule{Frequency: "weekly", Interval: 1},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodGet, "/tasks/"+fmt.Sprint(task.ID)+"/recurrence/next", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "2026-10-09T09:00:00Z") {
		t.Fatalf("unexpected next recurrence response: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestRecurringTaskWorkflowThroughRouter(t *testing.T) {
	service.ResetTasks()
	created := httptest.NewRecorder()
	payload := `{"title":"Daily review","dueDate":"2026-10-01T09:00:00Z","recurrence":{"frequency":"daily","interval":1}}`
	Router(created, httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(payload)))
	if created.Code != http.StatusCreated {
		t.Fatalf("create status %d: %s", created.Code, created.Body.String())
	}
	var task struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(created.Body).Decode(&task); err != nil {
		t.Fatal(err)
	}
	completed := httptest.NewRecorder()
	Router(completed, httptest.NewRequest(http.MethodPatch, "/tasks/"+fmt.Sprint(task.ID)+"/complete", nil))
	if completed.Code != http.StatusOK {
		t.Fatalf("complete status %d: %s", completed.Code, completed.Body.String())
	}
	history := httptest.NewRecorder()
	Router(history, httptest.NewRequest(http.MethodGet, "/tasks/"+fmt.Sprint(task.ID)+"/occurrences", nil))
	if history.Code != http.StatusOK || strings.Count(history.Body.String(), `"recurrenceOccurrence"`) != 2 {
		t.Fatalf("history response %d: %s", history.Code, history.Body.String())
	}
	preview := httptest.NewRecorder()
	Router(preview, httptest.NewRequest(http.MethodGet, "/tasks/"+fmt.Sprint(task.ID)+"/recurrence/next", nil))
	if preview.Code != http.StatusOK || !strings.Contains(preview.Body.String(), "2026-10-02T09:00:00Z") {
		t.Fatalf("preview response %d: %s", preview.Code, preview.Body.String())
	}
}

func TestActivityRouteSupportsPagination(t *testing.T) {
	service.ResetTasks()
	service.CreateTask(model.Task{Title: "First event"})
	service.CreateTask(model.Task{Title: "Second event"})
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodGet, "/activity?offset=1&limit=1", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"total":2`) || !strings.Contains(recorder.Body.String(), `"limit":1`) {
		t.Fatalf("unexpected paginated activity response: %d %s", recorder.Code, recorder.Body.String())
	}
	invalid := httptest.NewRecorder()
	Router(invalid, httptest.NewRequest(http.MethodGet, "/activity?limit=0", nil))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid pagination status %d, got %d", http.StatusBadRequest, invalid.Code)
	}
}

func TestActivityRouteFiltersByAction(t *testing.T) {
	service.ResetTasks()
	task := service.CreateTask(model.Task{Title: "Filtered task"})
	service.SetTaskCompletion(task.ID, true)
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodGet, "/tasks/"+fmt.Sprint(task.ID)+"/activity?action=completed", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"total":1`) || !strings.Contains(recorder.Body.String(), `"action":"completed"`) {
		t.Fatalf("unexpected action-filtered activity response: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestActivityRouteValidatesTimeWindow(t *testing.T) {
	service.ResetTasks()
	invalid := httptest.NewRecorder()
	Router(invalid, httptest.NewRequest(http.MethodGet, "/activity?from=not-a-date", nil))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected malformed timestamp status %d, got %d", http.StatusBadRequest, invalid.Code)
	}
	reversed := httptest.NewRecorder()
	Router(reversed, httptest.NewRequest(http.MethodGet, "/activity?from=2026-09-29T13:00:00Z&to=2026-09-29T12:00:00Z", nil))
	if reversed.Code != http.StatusBadRequest {
		t.Fatalf("expected reversed window status %d, got %d", http.StatusBadRequest, reversed.Code)
	}
}

func TestActivityRouteRejectsUnknownAction(t *testing.T) {
	service.ResetTasks()
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodGet, "/activity?action=made_up", nil))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected unknown action status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestDeletedTaskActivityRemainsAvailable(t *testing.T) {
	service.ResetTasks()
	task := service.CreateTask(model.Task{Title: "Archived task"})
	if !service.DeleteTask(task.ID) {
		t.Fatal("expected task deletion")
	}
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodGet, "/tasks/"+fmt.Sprint(task.ID)+"/activity", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"total":2`) || !strings.Contains(recorder.Body.String(), "Task deleted: Archived task") {
		t.Fatalf("deleted task activity was not retained: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestActivityReturnsEmptyPageForNoMatches(t *testing.T) {
	service.ResetTasks()
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodGet, "/activity?action=deleted", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"activities":[]`) || !strings.Contains(recorder.Body.String(), `"total":0`) {
		t.Fatalf("unexpected empty activity page: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestActivityTimelineAcrossTaskFeatures(t *testing.T) {
	service.ResetTasks()
	created := service.CreateTask(model.Task{Title: "Release checklist"})
	if _, err := service.AddTaskDependency(created.ID, 2); err != nil {
		t.Fatal(err)
	}
	item, err := service.AddChecklistItem(created.ID, "Publish notes")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetChecklistItemCompletion(created.ID, item.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetTaskCompletionChecked(2, true); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetTaskCompletionChecked(created.ID, true); err != nil {
		t.Fatal(err)
	}
	page := service.GetActivities(&created.ID, "", nil, nil, 0, 20)
	want := []string{"completed", "checklist_item_completed", "checklist_item_added", "dependency_added", "created"}
	if page.Total != len(want) || len(page.Activities) != len(want) {
		t.Fatalf("expected %d timeline events, got %#v", len(want), page)
	}
	for index, action := range want {
		if page.Activities[index].Action != action {
			t.Fatalf("event %d: expected action %q, got %#v", index, action, page.Activities[index])
		}
	}
}

func TestTaskDependencyWorkflowThroughRouter(t *testing.T) {
	service.ResetTasks()
	addRequest := httptest.NewRequest(http.MethodPost, "/tasks/1/dependencies", strings.NewReader(`{"dependsOn":2}`))
	addRecorder := httptest.NewRecorder()
	Router(addRecorder, addRequest)
	if addRecorder.Code != http.StatusCreated {
		t.Fatalf("expected dependency creation status %d, got %d: %s", http.StatusCreated, addRecorder.Code, addRecorder.Body.String())
	}

	readyRecorder := httptest.NewRecorder()
	Router(readyRecorder, httptest.NewRequest(http.MethodGet, "/tasks/ready", nil))
	if readyRecorder.Code != http.StatusOK || strings.Contains(readyRecorder.Body.String(), `"id":1`) {
		t.Fatalf("expected dependent task excluded from ready queue, got %d: %s", readyRecorder.Code, readyRecorder.Body.String())
	}

	prerequisiteRecorder := httptest.NewRecorder()
	Router(prerequisiteRecorder, httptest.NewRequest(http.MethodPatch, "/tasks/2/complete", nil))
	if prerequisiteRecorder.Code != http.StatusOK {
		t.Fatalf("expected prerequisite completion status %d, got %d", http.StatusOK, prerequisiteRecorder.Code)
	}
	dependentRecorder := httptest.NewRecorder()
	Router(dependentRecorder, httptest.NewRequest(http.MethodPatch, "/tasks/1/complete", nil))
	if dependentRecorder.Code != http.StatusOK {
		t.Fatalf("expected dependent completion status %d, got %d: %s", http.StatusOK, dependentRecorder.Code, dependentRecorder.Body.String())
	}
}

func TestDependenciesRouteIsDispatched(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks/1/dependencies", nil)
	recorder := httptest.NewRecorder()
	Router(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected dependency route status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestChecklistCollectionRouteSupportsGetAndPost(t *testing.T) {
	service.ResetTasks()
	getRecorder := httptest.NewRecorder()
	Router(getRecorder, httptest.NewRequest(http.MethodGet, "/tasks/1/checklist", nil))
	if getRecorder.Code != http.StatusOK || getRecorder.Body.String() != "[]\n" {
		t.Fatalf("unexpected checklist GET response: %d %s", getRecorder.Code, getRecorder.Body.String())
	}
	postRecorder := httptest.NewRecorder()
	Router(postRecorder, httptest.NewRequest(http.MethodPost, "/tasks/1/checklist", strings.NewReader(`{"text":"Review draft"}`)))
	if postRecorder.Code != http.StatusCreated || !strings.Contains(postRecorder.Body.String(), `"text":"Review draft"`) {
		t.Fatalf("unexpected checklist POST response: %d %s", postRecorder.Code, postRecorder.Body.String())
	}
}

func TestChecklistOrderRoute(t *testing.T) {
	service.ResetTasks()
	first, _ := service.AddChecklistItem(1, "first")
	second, _ := service.AddChecklistItem(1, "second")
	body := `{"ids":[` + fmt.Sprint(second.ID) + `,` + fmt.Sprint(first.ID) + `]}`
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodPut, "/tasks/1/checklist/order", strings.NewReader(body)))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"id":2`) {
		t.Fatalf("unexpected checklist order response: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestChecklistProgressRoute(t *testing.T) {
	service.ResetTasks()
	item, _ := service.AddChecklistItem(1, "finish")
	service.SetChecklistItemCompletion(1, item.ID, true)
	recorder := httptest.NewRecorder()
	Router(recorder, httptest.NewRequest(http.MethodGet, "/tasks/1/checklist/progress", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"percent":100`) {
		t.Fatalf("unexpected checklist progress response: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestChecklistLifecycleThroughRouter(t *testing.T) {
	service.ResetTasks()
	created := httptest.NewRecorder()
	Router(created, httptest.NewRequest(http.MethodPost, "/tasks/1/checklist", strings.NewReader(`{"text":"Ship feature"}`)))
	if created.Code != http.StatusCreated {
		t.Fatalf("create status %d: %s", created.Code, created.Body.String())
	}
	var item struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(created.Body).Decode(&item); err != nil {
		t.Fatal(err)
	}
	itemPath := "/tasks/1/checklist/" + fmt.Sprint(item.ID)
	patched := httptest.NewRecorder()
	Router(patched, httptest.NewRequest(http.MethodPatch, itemPath, strings.NewReader(`{"completed":true}`)))
	if patched.Code != http.StatusOK {
		t.Fatalf("patch status %d: %s", patched.Code, patched.Body.String())
	}
	progress := httptest.NewRecorder()
	Router(progress, httptest.NewRequest(http.MethodGet, "/tasks/1/checklist/progress", nil))
	if progress.Code != http.StatusOK || !strings.Contains(progress.Body.String(), `"percent":100`) {
		t.Fatalf("progress response %d: %s", progress.Code, progress.Body.String())
	}
	deleted := httptest.NewRecorder()
	Router(deleted, httptest.NewRequest(http.MethodDelete, itemPath, nil))
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status %d", deleted.Code)
	}
}
