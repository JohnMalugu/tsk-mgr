package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestHandleTasksGet(t *testing.T) {
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

func TestHandleTaskByIDRejectsInvalidID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/tasks/not-an-id", nil)
	recorder := httptest.NewRecorder()

	HandleTaskByID(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestHandleTaskByIDUpdatesTask(t *testing.T) {
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
