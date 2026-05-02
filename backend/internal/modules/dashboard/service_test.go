package dashboard

import "testing"

type testCounterRepo struct {
	count int
}

func (r testCounterRepo) Count() int {
	return r.count
}

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

func TestOverviewCountsFromCounterRepositories(t *testing.T) {
	service := NewService(
		testCounterRepo{count: 11},
		testCounterRepo{count: 13},
		testCounterRepo{count: 17},
	)

	overview := service.Overview()
	if overview.UserCount != 11 {
		t.Fatalf("expected user count 11, got %d", overview.UserCount)
	}
	if overview.ActivityCount != 13 {
		t.Fatalf("expected activity count 13, got %d", overview.ActivityCount)
	}
	if overview.ApplicationCount != 17 {
		t.Fatalf("expected application count 17, got %d", overview.ApplicationCount)
	}
}
