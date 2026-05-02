package service

import (
	"context"
	"testing"

	activityservice "credit-management/backend/internal/modules/activities/service"
	authservice "credit-management/backend/internal/modules/auth/service"
	participantservice "credit-management/backend/internal/modules/participants/service"
)

func TestApprovedActivityCreatesOrRestoresApplications(t *testing.T) {
	svc := New()

	effect := svc.SyncApprovedActivity("activity-1", []string{"u1", "u2"})
	if effect.AffectedCount != 2 {
		t.Fatalf("expected affected count 2, got %d", effect.AffectedCount)
	}
}

func TestSyncApprovedActivityContextCopiesParticipantCredits(t *testing.T) {
	ctx := context.Background()

	userStore, err := authservice.NewMemoryStore()
	if err != nil {
		t.Fatalf("new memory user store: %v", err)
	}

	activityRepo := activityservice.NewMemoryRepository()
	activitySvc := activityservice.New(activityRepo, nil)
	participantSvc := participantservice.New(activityRepo, userStore)

	activity, err := activitySvc.Create(ctx, "01JSTUDENT0000000000000000", "student", activityservice.Input{
		Title:        "Credit sync activity",
		Category:     "innovation",
		AcademicYear: "2025-2026",
		Semester:     "autumn",
		Details:      map[string]any{},
	})
	if err != nil {
		t.Fatalf("create activity: %v", err)
	}

	if _, err := participantSvc.Add(
		ctx,
		"01JSTUDENT0000000000000000",
		"student",
		activity.ID,
		[]string{"01JSTUDENT0000000000000000"},
		2.5,
	); err != nil {
		t.Fatalf("add participant: %v", err)
	}

	svc := New()
	svc.SetActivityRepository(activityRepo)
	svc.SetParticipantService(participantSvc)

	if err := svc.SyncApprovedActivityContext(ctx, activity.ID); err != nil {
		t.Fatalf("sync approved activity context: %v", err)
	}

	items := svc.List("01JSTUDENT0000000000000000", "student")
	if len(items) != 1 {
		t.Fatalf("expected 1 application, got %d", len(items))
	}
	if items[0].AppliedCredits != 2.5 || items[0].AwardedCredits != 2.5 {
		t.Fatalf("expected synced credits 2.5, got applied=%v awarded=%v", items[0].AppliedCredits, items[0].AwardedCredits)
	}

	stats := svc.Stats("01JSTUDENT0000000000000000", "student")
	if stats.TotalCredits != 2.5 {
		t.Fatalf("expected total credits 2.5, got %v", stats.TotalCredits)
	}
}
