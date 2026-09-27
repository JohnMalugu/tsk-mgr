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

func TestHandleTasksFiltersOverdue(t *testing.T) {
	resetTaskFixture()
	request := httptest.NewRequest(http.MethodGet, "/tasks?overdue=true", nil)
	recorder := httptest.NewRecorder()
	HandleTasks(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"id":1`) || strings.Contains(recorder.Body.String(), `"id":2`) {
		t.Fatalf("expected only overdue task 1, got %q", recorder.Body.String())
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

func TestHandleTasksRejectsTrailingJSON(t *testing.T) {
	resetTaskFixture()
	body := strings.NewReader(`{"title":"Task","dueDate":"2030-01-02T15:04:05Z"} {}`)
	request := httptest.NewRequest(http.MethodPost, "/tasks", body)
	recorder := httptest.NewRecorder()
	HandleTasks(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandleTasksRejectsOversizedBody(t *testing.T) {
	resetTaskFixture()
	body := `{"title":"` + strings.Repeat("a", 1<<20) + `","dueDate":"2030-01-02T15:04:05Z"}`
	request := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
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

func TestHandleTasksSearchesTags(t *testing.T) {
	resetTaskFixture()
	request := httptest.NewRequest(http.MethodGet, "/tasks?q=errands", nil)
	recorder := httptest.NewRecorder()
	HandleTasks(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "Buy groceries") || strings.Contains(recorder.Body.String(), "Learn Go") {
		t.Fatalf("expected search to match the errands tag only, got %q", recorder.Body.String())
	}
}

func TestHandleTasksFiltersByDueDateRange(t *testing.T) {
	resetTaskFixture()
	for _, payload := range []string{
		`{"title":"June task","dueDate":"2030-06-15T12:00:00Z"}`,
		`{"title":"July task","dueDate":"2030-07-15T12:00:00Z"}`,
	} {
		recorder := httptest.NewRecorder()
		HandleTasks(recorder, httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(payload)))
		if recorder.Code != http.StatusCreated {
			t.Fatalf("expected task creation status %d, got %d", http.StatusCreated, recorder.Code)
		}
	}
	request := httptest.NewRequest(http.MethodGet, "/tasks?dueAfter=2030-06-01T00:00:00Z&dueBefore=2030-06-30T23:59:59Z", nil)
	recorder := httptest.NewRecorder()
	HandleTasks(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "June task") || strings.Contains(recorder.Body.String(), "July task") {
		t.Fatalf("expected only June task in range, got %q", recorder.Body.String())
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
	request := httptest.NewRequest(http.MethodGet, "/tasks?sort=unknown", nil)
	recorder := httptest.NewRecorder()

	HandleTasks(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandleTasksSortsByPriority(t *testing.T) {
	resetTaskFixture()
	body := bytes.NewBufferString(`{"title":"Urgent task","dueDate":"2030-01-02T15:04:05Z","priority":"high"}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/tasks", body)
	createRecorder := httptest.NewRecorder()
	HandleTasks(createRecorder, createRequest)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected task creation status %d, got %d", http.StatusCreated, createRecorder.Code)
	}

	for _, test := range []struct {
		order string
		want  []int
	}{
		{order: "asc", want: []int{2, 1, 3}},
		{order: "desc", want: []int{3, 1, 2}},
	} {
		request := httptest.NewRequest(http.MethodGet, "/tasks?sort=priority&order="+test.order, nil)
		recorder := httptest.NewRecorder()
		HandleTasks(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("order %s: expected status %d, got %d", test.order, http.StatusOK, recorder.Code)
		}

		var tasks []struct {
			ID int `json:"id"`
		}
		if err := json.NewDecoder(recorder.Body).Decode(&tasks); err != nil {
			t.Fatalf("decode priority-sorted tasks: %v", err)
		}
		if len(tasks) != len(test.want) {
			t.Fatalf("order %s: expected %d tasks, got %#v", test.order, len(test.want), tasks)
		}
		for index, task := range tasks {
			if task.ID != test.want[index] {
				t.Fatalf("order %s: expected task IDs %v, got %#v", test.order, test.want, tasks)
			}
		}
	}
}

func TestHandleTasksSortsByCompletion(t *testing.T) {
	resetTaskFixture()
	service.SetTaskCompletion(2, true)
	request := httptest.NewRequest(http.MethodGet, "/tasks?sort=completed&order=desc", nil)
	recorder := httptest.NewRecorder()
	HandleTasks(recorder, request)
	var tasks []struct {
		ID        int  `json:"id"`
		Completed bool `json:"completed"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&tasks); err != nil {
		t.Fatalf("decode completion-sorted tasks: %v", err)
	}
	if len(tasks) != 2 || !tasks[0].Completed || tasks[0].ID != 2 {
		t.Fatalf("expected completed tasks first, got %#v", tasks)
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

func TestHandleTasksFiltersByTag(t *testing.T) {
	resetTaskFixture()
	body := bytes.NewBufferString(`{"title":"Errand run","dueDate":"2030-01-02T15:04:05Z","tags":["home","errands"]}`)
	createRequest := httptest.NewRequest(http.MethodPost, "/tasks", body)
	createRecorder := httptest.NewRecorder()
	HandleTasks(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRecorder.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/tasks?tag=home", nil)
	recorder := httptest.NewRecorder()
	HandleTasks(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	var tasks []struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&tasks); err != nil {
		t.Fatalf("decode tagged tasks: %v", err)
	}
	if len(tasks) == 0 {
		t.Fatal("expected at least one task with the requested tag")
	}
	for _, task := range tasks {
		if !contains(task.Tags, "home") {
			t.Fatalf("expected only home-tagged tasks, got %#v", tasks)
		}
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
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

func TestHandleBulkCompleteMarksTasksCompleted(t *testing.T) {
	resetTaskFixture()
	body := bytes.NewBufferString(`{"ids":[1,2]}`)
	request := httptest.NewRequest(http.MethodPost, "/tasks/bulk/complete", body)
	recorder := httptest.NewRecorder()

	HandleBulkComplete(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"updated":2`) {
		t.Fatalf("expected two updated tasks, got %q", recorder.Body.String())
	}
}

func TestHandleBulkCompleteRejectsMissingTask(t *testing.T) {
	resetTaskFixture()
	body := bytes.NewBufferString(`{"ids":[1,999]}`)
	request := httptest.NewRequest(http.MethodPost, "/tasks/bulk/complete", body)
	recorder := httptest.NewRecorder()

	HandleBulkComplete(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestHandleBulkCompleteRejectsEmptyIDs(t *testing.T) {
	body := bytes.NewBufferString(`{"ids":[]}`)
	request := httptest.NewRequest(http.MethodPost, "/tasks/bulk/complete", body)
	recorder := httptest.NewRecorder()

	HandleBulkComplete(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
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
		Total      int            `json:"total"`
		Completed  int            `json:"completed"`
		Pending    int            `json:"pending"`
		Overdue    int            `json:"overdue"`
		Blocked    int            `json:"blocked"`
		ByPriority map[string]int `json:"byPriority"`
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
	if summary.ByPriority["medium"] != 1 || summary.ByPriority["low"] != 1 || summary.ByPriority["high"] != 0 {
		t.Fatalf("unexpected priority counts: %#v", summary.ByPriority)
	}
}

func TestHandleTasksSummaryCountsBlockedTasks(t *testing.T) {
	resetTaskFixture()
	if _, err := service.AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	HandleTaskSummary(recorder, httptest.NewRequest(http.MethodGet, "/tasks/summary", nil))
	var summary struct {
		Blocked int `json:"blocked"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&summary); err != nil {
		t.Fatal(err)
	}
	if summary.Blocked != 1 {
		t.Fatalf("expected one blocked task, got %d", summary.Blocked)
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

func TestHandleTaskByIDPatchesTask(t *testing.T) {
	resetTaskFixture()
	request := httptest.NewRequest(http.MethodPatch, "/tasks/1", strings.NewReader(`{"description":"Bring reusable bags"}`))
	recorder := httptest.NewRecorder()
	HandleTaskByID(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	var task struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&task); err != nil {
		t.Fatalf("decode patched task: %v", err)
	}
	if task.Title != "Buy groceries" || task.Description != "Bring reusable bags" {
		t.Fatalf("expected description-only patch, got %#v", task)
	}
}

func TestHandleTaskDependenciesAddsPrerequisite(t *testing.T) {
	resetTaskFixture()
	body := strings.NewReader(`{"dependsOn":2}`)
	request := httptest.NewRequest(http.MethodPost, "/tasks/1/dependencies", body)
	recorder := httptest.NewRecorder()
	HandleTaskDependencies(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"dependsOn":[2]`) {
		t.Fatalf("expected updated dependency IDs, got %q", recorder.Body.String())
	}
}

func TestCreateTaskWithPrerequisiteIDs(t *testing.T) {
	resetTaskFixture()
	body := strings.NewReader(`{"title":"Prepare report","dueDate":"2030-01-02T15:04:05Z","dependsOn":[2]}`)
	recorder := httptest.NewRecorder()
	HandleTasks(recorder, httptest.NewRequest(http.MethodPost, "/tasks", body))
	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"dependsOn":[2]`) {
		t.Fatalf("expected dependency IDs in response, got %q", recorder.Body.String())
	}
}

func TestCreateTaskRejectsUnknownPrerequisite(t *testing.T) {
	resetTaskFixture()
	body := strings.NewReader(`{"title":"Prepare report","dueDate":"2030-01-02T15:04:05Z","dependsOn":[999]}`)
	recorder := httptest.NewRecorder()
	HandleTasks(recorder, httptest.NewRequest(http.MethodPost, "/tasks", body))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
	if len(service.GetAllTasks()) != 2 {
		t.Fatal("invalid prerequisite reference must not create a task")
	}
}

func TestHandleReadyTasksExcludesBlockedTasks(t *testing.T) {
	resetTaskFixture()
	if _, err := service.AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/tasks/ready", nil)
	recorder := httptest.NewRecorder()
	HandleReadyTasks(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), `"id":1`) || !strings.Contains(recorder.Body.String(), `"id":2`) {
		t.Fatalf("expected only ready prerequisite task, got %q", recorder.Body.String())
	}
}

func TestHandleTaskDependencyDeletesPrerequisite(t *testing.T) {
	resetTaskFixture()
	if _, err := service.AddTaskDependency(1, 2); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodDelete, "/tasks/1/dependencies/2", nil)
	recorder := httptest.NewRecorder()
	HandleTaskDependency(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
	dependencies, found := service.GetTaskDependencies(1)
	if !found || len(dependencies) != 0 {
		t.Fatalf("expected prerequisite removed, got %#v", dependencies)
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
	request := httptest.NewRequest(http.MethodPost, "/tasks/1", nil)
	recorder := httptest.NewRecorder()

	HandleTaskByID(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, recorder.Code)
	}
	if allow := recorder.Header().Get("Allow"); allow != "GET, PUT, PATCH, DELETE" {
		t.Fatalf("expected Allow header %q, got %q", "GET, PUT, PATCH, DELETE", allow)
	}
}
