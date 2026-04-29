package domain

import (
	"errors"
	"time"
)

const (
	StatusDraft         = "draft"
	StatusPendingReview = "pending_review"
	StatusApproved      = "approved"
	StatusRejected      = "rejected"
)

var (
	ErrActivityStatusInvalid = errors.New("activity status invalid")
	ErrActivityForbidden     = errors.New("activity forbidden")
	ErrActivityNotFound      = errors.New("activity not found")
)

type Activity struct {
	ID             string
	Title          string
	Description    string
	Category       string
	AcademicYear   string
	Semester       string
	StartDate      string
	EndDate        string
	Status         string
	OwnerID        string
	ReviewerID     string
	ReviewComments string
	ReviewedAt     *time.Time
	Details        map[string]any
	DeletedAt      *time.Time
}

func (a Activity) AllowStructuralEdit() error {
	if a.Status != StatusDraft {
		return ErrActivityStatusInvalid
	}

	return nil
}

func (a Activity) CanDelete(actorID string, actorRole string) bool {
	if actorRole == "teacher" || actorRole == "admin" {
		return true
	}

	return actorID == a.OwnerID && a.Status == StatusDraft
}

func (a Activity) CanView(actorID string, actorRole string) bool {
	if actorRole == "teacher" || actorRole == "admin" {
		return true
	}

	return actorID == a.OwnerID
}

func (a *Activity) Submit(actorID string) error {
	if actorID != a.OwnerID {
		return ErrActivityForbidden
	}
	if a.Status == StatusPendingReview {
		return nil
	}
	if a.Status != StatusDraft {
		return ErrActivityStatusInvalid
	}

	a.Status = StatusPendingReview
	return nil
}

func (a *Activity) Withdraw(actorID string) error {
	if actorID != a.OwnerID {
		return ErrActivityForbidden
	}
	if a.Status == StatusDraft {
		return nil
	}
	if a.Status != StatusPendingReview && a.Status != StatusRejected {
		return ErrActivityStatusInvalid
	}

	a.Status = StatusDraft
	return nil
}

func (a *Activity) Review(actorRole string, targetStatus string, reviewComments string, reviewerID string, reviewedAt time.Time) error {
	if actorRole != "teacher" && actorRole != "admin" {
		return ErrActivityForbidden
	}
	if targetStatus != StatusApproved && targetStatus != StatusRejected {
		return ErrActivityStatusInvalid
	}
	if a.Status == targetStatus {
		a.ReviewComments = reviewComments
		a.ReviewerID = reviewerID
		a.ReviewedAt = &reviewedAt
		return nil
	}
	if a.Status != StatusPendingReview && !(a.Status == StatusApproved && targetStatus == StatusRejected) && !(a.Status == StatusRejected && targetStatus == StatusApproved) {
		return ErrActivityStatusInvalid
	}

	a.Status = targetStatus
	a.ReviewComments = reviewComments
	a.ReviewerID = reviewerID
	a.ReviewedAt = &reviewedAt
	return nil
}

func (a *Activity) TransferOwner(actorRole string, newOwnerID string) error {
	if actorRole != "teacher" && actorRole != "admin" {
		return ErrActivityForbidden
	}
	if newOwnerID == "" {
		return ErrActivityStatusInvalid
	}

	a.OwnerID = newOwnerID
	return nil
}
