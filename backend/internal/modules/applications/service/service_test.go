package service

import (
	"bytes"
	"context"
	"testing"

	activityservice "credit-management/backend/internal/modules/activities/service"
	authservice "credit-management/backend/internal/modules/auth/service"
	participantservice "credit-management/backend/internal/modules/participants/service"
	"credit-management/backend/internal/platform/exportfile"

	"github.com/xuri/excelize/v2"
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

func TestListFilteredPagePaginatesAfterFiltering(t *testing.T) {
	svc := NewWithRepository(fakeApplicationRepository{
		items: []Application{
			{ID: "app-1", ActivityID: "activity-1", UserID: "student-1", Status: "approved"},
			{ID: "app-2", ActivityID: "activity-2", UserID: "student-2", Status: "rejected"},
			{ID: "app-3", ActivityID: "activity-3", UserID: "student-3", Status: "approved"},
			{ID: "app-4", ActivityID: "activity-4", UserID: "student-4", Status: "approved"},
		},
	})

	result := svc.ListFilteredPage("admin-1", "admin", Filter{
		Status:   "approved",
		Page:     2,
		PageSize: 2,
	})

	if result.Total != 3 || result.Page != 2 || result.PageSize != 2 || result.TotalPages != 2 {
		t.Fatalf("unexpected pagination metadata: %#v", result)
	}
	if len(result.Items) != 1 || result.Items[0].ID != "app-4" {
		t.Fatalf("unexpected paginated items: %#v", result.Items)
	}
}

func TestExportFilteredSupportsXLSX(t *testing.T) {
	svc := NewWithRepository(fakeApplicationRepository{
		items: []Application{
			{
				ID:             "app-1",
				ActivityID:     "activity-1",
				UserID:         "student-1",
				Status:         "approved",
				AppliedCredits: 2.5,
				AwardedCredits: 2,
			},
		},
	})

	file, err := svc.ExportFiltered("admin-1", "admin", Filter{}, exportfile.FormatXLSX)
	if err != nil {
		t.Fatalf("ExportFiltered returned error: %v", err)
	}
	if file.Filename != "applications.xlsx" || file.ContentType != exportfile.ContentTypeXLSX {
		t.Fatalf("unexpected file metadata: %#v", file)
	}

	workbook, err := excelize.OpenReader(bytes.NewReader(file.Content))
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	defer func() {
		if closeErr := workbook.Close(); closeErr != nil {
			t.Fatalf("close workbook: %v", closeErr)
		}
	}()

	rows, err := workbook.GetRows("Applications")
	if err != nil {
		t.Fatalf("read xlsx rows: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "id" || rows[1][0] != "app-1" {
		t.Fatalf("unexpected xlsx rows: %#v", rows)
	}
}

type fakeApplicationRepository struct {
	items []Application
}

func (f fakeApplicationRepository) UpsertApproved(context.Context, Application) error {
	return nil
}

func (f fakeApplicationRepository) SoftDeleteByActivity(context.Context, string) error {
	return nil
}

func (f fakeApplicationRepository) List(_ context.Context, actorID string, actorRole string) ([]Application, error) {
	items := make([]Application, 0, len(f.items))
	for _, item := range f.items {
		if item.Deleted {
			continue
		}
		if actorRole == "student" && item.UserID != actorID {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (f fakeApplicationRepository) Get(_ context.Context, actorID string, actorRole string, applicationID string) (*Application, error) {
	for _, item := range f.items {
		if item.ID != applicationID || item.Deleted {
			continue
		}
		if actorRole == "student" && item.UserID != actorID {
			break
		}
		copy := item
		return &copy, nil
	}
	return nil, ErrApplicationNotFound
}

func (f fakeApplicationRepository) Count(context.Context) (int, error) {
	count := 0
	for _, item := range f.items {
		if !item.Deleted {
			count++
		}
	}
	return count, nil
}
