package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

func resetTaskFixture() {
	service.ResetTasks()
}

func TestHandleTasksGet(t *testing.T) {
	resetTaskFixture()
	request := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}
	if !strings.Contains(recorder.Body.String(), "Buy groceries") {
		t.Fatalf("expected seeded task in response, got %q", recorder.Body.String())
	}
}

func TestHandleTasksFiltersByCompletion(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks?completed=true", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	var tasks []struct {
		Completed bool `json:"completed"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&tasks); err != nil {
		t.Fatalf("decode filtered tasks: %v", err)
	}
	for _, task := range tasks {
		if !task.Completed {
			t.Fatal("expected only completed tasks")
		}
	}
}

func TestHandleTasksRejectsInvalidCompletionFilter(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks?completed=maybe", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandleTasksFiltersByTitle(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks?q=grocer", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "Buy groceries") {
		t.Fatalf("expected matching task in response, got %q", recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "Learn Go") {
		t.Fatalf("expected non-matching task to be excluded, got %q", recorder.Body.String())
	}
}

func TestHandleTasksSortsByTitleDescending(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks?sort=title&order=desc", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	var tasks []struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&tasks); err != nil {
		t.Fatalf("decode sorted tasks: %v", err)
	}
	if len(tasks) < 2 || tasks[0].Title != "Learn Go" {
		t.Fatalf("expected Learn Go first, got %#v", tasks)
	}
}

func TestHandleTasksSortsByDueDateAscending(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks?sort=dueDate", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	var tasks []struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&tasks); err != nil {
		t.Fatalf("decode due-date sorted tasks: %v", err)
	}
	if len(tasks) < 2 || tasks[0].ID != 1 {
		t.Fatalf("expected overdue task first, got %#v", tasks)
	}
}

func TestHandleTasksRejectsInvalidSort(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks?sort=priority", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandleTasksRejectsInvalidOrder(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks?order=random", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandleTasksFiltersByPriority(t *testing.T) {
	resetTaskFixture()
	body := bytes.NewBufferString(`{"title":"Critical task","dueDate":"2030-01-02T15:04:05Z","priority":"high"}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/tasks", body)
	createRecorder := httptest.NewRecorder()
	HandleTasks(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRecorder.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/tasks?priority=high", nil)
	recorder := httptest.NewRecorder()
	HandleTasks(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	var tasks []struct {
		Priority string `json:"priority"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&tasks); err != nil {
		t.Fatalf("decode priority-filtered tasks: %v", err)
	}
	for _, task := range tasks {
		if task.Priority != "high" {
			t.Fatalf("expected only high priority tasks, got %#v", tasks)
		}
	}
}

func TestHandleTasksRejectsInvalidPriority(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks?priority=urgent", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandleTasksPaginatesResults(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks?offset=1&limit=1", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	var tasks []struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&tasks); err != nil {
		t.Fatalf("decode paginated tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != 2 {
		t.Fatalf("expected task 2, got %#v", tasks)
	}
}

func TestHandleTasksRejectsInvalidPagination(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks?limit=101", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandleTasksRejectsNegativeOffset(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks?offset=-1", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandleTaskCompleteMarksCompleted(t *testing.T) {
	resetTaskFixture()
	request := httptest.NewRequest(http.MethodPatch, "/tasks/1/complete", nil)
	recorder := httptest.NewRecorder()

	HandleTaskComplete(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"completed":true`) {
		t.Fatalf("expected completed task in response, got %q", recorder.Body.String())
	}
}

func TestHandleTaskCompleteRejectsMissingTask(t *testing.T) {
	resetTaskFixture()
	request := httptest.NewRequest(http.MethodPatch, "/tasks/999/complete", nil)
	recorder := httptest.NewRecorder()

	HandleTaskComplete(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestHandleTaskCompleteRejectsInvalidJSON(t *testing.T) {
	resetTaskFixture()
	body := bytes.NewBufferString(`{"completed":"yes"}`)
	request := httptest.NewRequest(http.MethodPatch, "/tasks/1/complete", body)
	recorder := httptest.NewRecorder()

	HandleTaskComplete(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandleTasksSummaryReturnsCounts(t *testing.T) {
	resetTaskFixture()
	request := httptest.NewRequest(http.MethodGet, "/tasks/summary", nil)
	recorder := httptest.NewRecorder()

	HandleTaskSummary(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var summary struct {
		Total     int `json:"total"`
		Completed int `json:"completed"`
		Pending   int `json:"pending"`
		Overdue   int `json:"overdue"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&summary); err != nil {
		t.Fatalf("decode task summary: %v", err)
	}
	if summary.Total != 2 {
		t.Fatalf("expected total 2, got %d", summary.Total)
	}
	if summary.Completed != 0 {
		t.Fatalf("expected completed 0, got %d", summary.Completed)
	}
	if summary.Pending != 2 {
		t.Fatalf("expected pending 2, got %d", summary.Pending)
	}
	if summary.Overdue != 1 {
		t.Fatalf("expected overdue 1, got %d", summary.Overdue)
	}
}

func TestHandleTaskByIDRejectsInvalidID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks/not-an-id", nil)
	recorder := httptest.NewRecorder()

	HandleTaskByID(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandleTaskByIDUpdatesTask(t *testing.T) {
	resetTaskFixture()
	body := bytes.NewBufferString(`{"title":"Updated task","dueDate":"2030-01-02T15:04:05Z","completed":true}`)
	request := httptest.NewRequest(http.MethodPut, "/tasks/2", body)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	HandleTaskByID(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"title":"Updated task"`) {
		t.Fatalf("expected updated task in response, got %q", recorder.Body.String())
	}
}

func TestHandleTaskByIDDeletesTask(t *testing.T) {
	resetTaskFixture()
	body := bytes.NewBufferString(`{"title":"Temporary task","dueDate":"2030-01-02T15:04:05Z"}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/tasks", body)
	createRecorder := httptest.NewRecorder()
	HandleTasks(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected task creation status %d, got %d", http.StatusCreated, createRecorder.Code)
	}

	var task struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(createRecorder.Body).Decode(&task); err != nil {
		t.Fatalf("decode created task: %v", err)
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/tasks/"+strconv.Itoa(task.ID), nil)
	deleteRecorder := httptest.NewRecorder()
	HandleTaskByID(deleteRecorder, deleteRequest)

	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRecorder.Code)
	}
	if deleteRecorder.Body.Len() != 0 {
		t.Fatalf("expected empty delete response, got %q", deleteRecorder.Body.String())
	}
}

func TestHandleTaskByIDRejectsUnsupportedMethod(t *testing.T) {
	request := httptest.NewRequest(http.MethodPatch, "/tasks/1", nil)
	recorder := httptest.NewRecorder()

	HandleTaskByID(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, recorder.Code)
	}
	if allow := recorder.Header().Get("Allow"); allow != "GET, PUT, DELETE" {
		t.Fatalf("expected Allow header %q, got %q", "GET, PUT, DELETE", allow)
	}
}
