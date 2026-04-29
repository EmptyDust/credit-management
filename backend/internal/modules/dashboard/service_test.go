package dashboard

import "testing"

func TestOverviewCounts(t *testing.T) {
	service := NewService(
		func() int { return 3 },
		func() int { return 5 },
		func() int { return 7 },
	)

	overview := service.Overview()
	if overview.UserCount != 3 {
		t.Fatalf("expected user count 3, got %d", overview.UserCount)
	}
	if overview.ActivityCount != 5 {
		t.Fatalf("expected activity count 5, got %d", overview.ActivityCount)
	}
	if overview.ApplicationCount != 7 {
		t.Fatalf("expected application count 7, got %d", overview.ApplicationCount)
	}
}
