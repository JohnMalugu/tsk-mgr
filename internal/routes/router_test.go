package routes

import (
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
