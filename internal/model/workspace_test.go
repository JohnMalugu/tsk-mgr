package model

import "testing"

func TestWorkspace(t *testing.T) {
	w := Workspace{Name: "My Workspace"}
	if w.Name != "My Workspace" {
		t.Errorf("expected My Workspace")
	}
}
