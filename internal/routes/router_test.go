package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
