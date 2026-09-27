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
