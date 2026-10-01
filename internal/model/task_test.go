package model

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
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

func TestTaskEstimateSerializesInMinutes(t *testing.T) {
	task := Task{ID: 4, EstimateMinutes: 90}
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Task
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.EstimateMinutes != 90 {
		t.Fatalf("expected 90 minute estimate, got %d", decoded.EstimateMinutes)
	}
}

func TestTaskRecurrenceRuleRoundTripsAsJSON(t *testing.T) {
	until := time.Date(2027, 1, 31, 10, 0, 0, 0, time.UTC)
	want := Task{
		ID:                   9,
		Recurrence:           &RecurrenceRule{Frequency: "monthly", Interval: 1, Until: &until},
		RecurrenceSeriesID:   9,
		RecurrenceOccurrence: 2,
	}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got Task
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Recurrence == nil || got.Recurrence.Frequency != "monthly" || got.Recurrence.Interval != 1 || got.Recurrence.Until == nil || !got.Recurrence.Until.Equal(until) || got.RecurrenceSeriesID != 9 || got.RecurrenceOccurrence != 2 {
		t.Fatalf("unexpected recurrence round trip: %#v", got)
	}
}
