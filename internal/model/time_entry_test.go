package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeEntryJSONRoundTrip(t *testing.T) {
	ended := time.Date(2026, 9, 30, 11, 0, 0, 0, time.UTC)
	want := TimeEntry{ID: 2, TaskID: 7, StartedAt: ended.Add(-time.Hour), EndedAt: &ended, DurationSeconds: 3600, Note: "Planning"}
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal time entry: %v", err)
	}
	var got TimeEntry
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal time entry: %v", err)
	}
	if got.ID != want.ID || got.TaskID != want.TaskID || got.DurationSeconds != want.DurationSeconds || got.Note != want.Note || got.EndedAt == nil || !got.EndedAt.Equal(ended) {
		t.Fatalf("unexpected time entry round trip: %#v", got)
	}
}
