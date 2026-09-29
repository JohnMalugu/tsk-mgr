package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestActivityJSONRoundTrip(t *testing.T) {
	want := Activity{ID: 8, TaskID: 3, Action: "completed", Summary: "Task completed", OccurredAt: time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal activity: %v", err)
	}
	var got Activity
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal activity: %v", err)
	}
	if got != want {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}
