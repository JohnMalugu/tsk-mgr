package model

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTaskDependencyIDsRoundTripAsJSON(t *testing.T) {
	task := Task{ID: 7, DependsOn: []int{2, 4}}
	encoded, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal task: %v", err)
	}
	var decoded Task
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal task: %v", err)
	}
	if !reflect.DeepEqual(decoded.DependsOn, task.DependsOn) {
		t.Fatalf("expected dependencies %v, got %v", task.DependsOn, decoded.DependsOn)
	}
}

func TestTaskChecklistRoundTripsAsJSON(t *testing.T) {
	task := Task{ID: 3, Checklist: []ChecklistItem{{ID: 1, Text: "Draft", Completed: true}}}
	encoded, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal task: %v", err)
	}
	var decoded Task
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal task: %v", err)
	}
	if len(decoded.Checklist) != 1 || decoded.Checklist[0] != task.Checklist[0] {
		t.Fatalf("expected checklist item round trip, got %#v", decoded.Checklist)
	}
}
