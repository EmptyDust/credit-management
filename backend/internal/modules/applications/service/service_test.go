package service

import "testing"

func TestApprovedActivityCreatesOrRestoresApplications(t *testing.T) {
	svc := New()

	effect := svc.SyncApprovedActivity("activity-1", []string{"u1", "u2"})
	if effect.AffectedCount != 2 {
		t.Fatalf("expected affected count 2, got %d", effect.AffectedCount)
	}
}
