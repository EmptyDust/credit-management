package service

import (
	"errors"
	"testing"

	activitydomain "credit-management/backend/internal/modules/activities/domain"
)

func TestUpdateDraftOnly(t *testing.T) {
	activity := activitydomain.Activity{Status: activitydomain.StatusApproved}

	err := activity.AllowStructuralEdit()
	if !errors.Is(err, activitydomain.ErrActivityStatusInvalid) {
		t.Fatalf("expected ErrActivityStatusInvalid, got %v", err)
	}
}
