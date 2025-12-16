package testutils

import (
	"time"

	"gorm.io/gorm"
)

// User represents a test user fixture
type User struct {
	ID         string `gorm:"primaryKey"`
	Username   string `gorm:"unique;not null"`
	Name       string
	Email      string `gorm:"unique"`
	Department string
	Role       string
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Activity represents a test activity fixture
type Activity struct {
	ID                string `gorm:"primaryKey"`
	Name              string `gorm:"not null"`
	Type              string
	Description       string
	OrganizerId       string
	StartTime         time.Time
	EndTime           time.Time
	Location          string
	MaxParticipants   int
	RegistrationStart time.Time
	RegistrationEnd   time.Time
	BaseCredit        float64
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ActivityParticipant represents a test participant fixture
type ActivityParticipant struct {
	ID           string `gorm:"primaryKey"`
	ActivityID   string `gorm:"not null;index"`
	UserID       string `gorm:"not null;index"`
	Status       string
	RegisteredAt time.Time
	AttendedAt   *time.Time
	CreditEarned float64
	Feedback     string
	Rating       *int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Department represents a test department fixture
type Department struct {
	ID        string `gorm:"primaryKey"`
	Name      string `gorm:"unique;not null"`
	Code      string `gorm:"unique"`
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Attachment represents a test attachment fixture
type Attachment struct {
	ID           string `gorm:"primaryKey"`
	ActivityID   string `gorm:"not null;index"`
	Filename     string
	OriginalName string
	FilePath     string
	FileSize     int64
	ContentType  string
	UploadedBy   string
	CreatedAt    time.Time
}

// FixtureBuilder helps create test data
type FixtureBuilder struct {
	db *gorm.DB
}

// NewFixtureBuilder creates a new fixture builder
func NewFixtureBuilder(db *gorm.DB) *FixtureBuilder {
	return &FixtureBuilder{db: db}
}

// CreateUser creates a test user with default or custom values
func (fb *FixtureBuilder) CreateUser(overrides map[string]interface{}) (*User, error) {
	user := &User{
		ID:         GenerateID(),
		Username:   "testuser_" + GenerateID()[:8],
		Name:       "Test User",
		Email:      "test_" + GenerateID()[:8] + "@example.com",
		Department: "Computer Science",
		Role:       "student",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Apply overrides
	if overrides != nil {
		if v, ok := overrides["id"].(string); ok {
			user.ID = v
		}
		if v, ok := overrides["username"].(string); ok {
			user.Username = v
		}
		if v, ok := overrides["name"].(string); ok {
			user.Name = v
		}
		if v, ok := overrides["email"].(string); ok {
			user.Email = v
		}
		if v, ok := overrides["department"].(string); ok {
			user.Department = v
		}
		if v, ok := overrides["role"].(string); ok {
			user.Role = v
		}
		if v, ok := overrides["is_active"].(bool); ok {
			user.IsActive = v
		}
	}

	if err := fb.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// CreateActivity creates a test activity with default or custom values
func (fb *FixtureBuilder) CreateActivity(overrides map[string]interface{}) (*Activity, error) {
	now := time.Now()
	activity := &Activity{
		ID:                GenerateID(),
		Name:              "Test Activity " + GenerateID()[:8],
		Type:              "workshop",
		Description:       "A test activity for testing purposes",
		OrganizerId:       "",
		StartTime:         now.Add(24 * time.Hour),
		EndTime:           now.Add(26 * time.Hour),
		Location:          "Test Location",
		MaxParticipants:   50,
		RegistrationStart: now,
		RegistrationEnd:   now.Add(12 * time.Hour),
		BaseCredit:        2.0,
		Status:            "draft",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	// Apply overrides
	if overrides != nil {
		if v, ok := overrides["id"].(string); ok {
			activity.ID = v
		}
		if v, ok := overrides["name"].(string); ok {
			activity.Name = v
		}
		if v, ok := overrides["type"].(string); ok {
			activity.Type = v
		}
		if v, ok := overrides["description"].(string); ok {
			activity.Description = v
		}
		if v, ok := overrides["organizer_id"].(string); ok {
			activity.OrganizerId = v
		}
		if v, ok := overrides["start_time"].(time.Time); ok {
			activity.StartTime = v
		}
		if v, ok := overrides["end_time"].(time.Time); ok {
			activity.EndTime = v
		}
		if v, ok := overrides["location"].(string); ok {
			activity.Location = v
		}
		if v, ok := overrides["max_participants"].(int); ok {
			activity.MaxParticipants = v
		}
		if v, ok := overrides["registration_start"].(time.Time); ok {
			activity.RegistrationStart = v
		}
		if v, ok := overrides["registration_end"].(time.Time); ok {
			activity.RegistrationEnd = v
		}
		if v, ok := overrides["base_credit"].(float64); ok {
			activity.BaseCredit = v
		}
		if v, ok := overrides["status"].(string); ok {
			activity.Status = v
		}
	}

	if err := fb.db.Create(activity).Error; err != nil {
		return nil, err
	}

	return activity, nil
}

// CreateParticipant creates a test participant with default or custom values
func (fb *FixtureBuilder) CreateParticipant(activityID, userID string, overrides map[string]interface{}) (*ActivityParticipant, error) {
	participant := &ActivityParticipant{
		ID:           GenerateID(),
		ActivityID:   activityID,
		UserID:       userID,
		Status:       "registered",
		RegisteredAt: time.Now(),
		CreditEarned: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Apply overrides
	if overrides != nil {
		if v, ok := overrides["id"].(string); ok {
			participant.ID = v
		}
		if v, ok := overrides["status"].(string); ok {
			participant.Status = v
		}
		if v, ok := overrides["registered_at"].(time.Time); ok {
			participant.RegisteredAt = v
		}
		if v, ok := overrides["attended_at"].(*time.Time); ok {
			participant.AttendedAt = v
		}
		if v, ok := overrides["credit_earned"].(float64); ok {
			participant.CreditEarned = v
		}
		if v, ok := overrides["feedback"].(string); ok {
			participant.Feedback = v
		}
		if v, ok := overrides["rating"].(*int); ok {
			participant.Rating = v
		}
	}

	if err := fb.db.Create(participant).Error; err != nil {
		return nil, err
	}

	return participant, nil
}

// CreateDepartment creates a test department with default or custom values
func (fb *FixtureBuilder) CreateDepartment(overrides map[string]interface{}) (*Department, error) {
	dept := &Department{
		ID:        GenerateID(),
		Name:      "Test Department " + GenerateID()[:8],
		Code:      "DEPT" + GenerateID()[:4],
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Apply overrides
	if overrides != nil {
		if v, ok := overrides["id"].(string); ok {
			dept.ID = v
		}
		if v, ok := overrides["name"].(string); ok {
			dept.Name = v
		}
		if v, ok := overrides["code"].(string); ok {
			dept.Code = v
		}
		if v, ok := overrides["is_active"].(bool); ok {
			dept.IsActive = v
		}
	}

	if err := fb.db.Create(dept).Error; err != nil {
		return nil, err
	}

	return dept, nil
}

// CreateAttachment creates a test attachment with default or custom values
func (fb *FixtureBuilder) CreateAttachment(activityID string, overrides map[string]interface{}) (*Attachment, error) {
	attachment := &Attachment{
		ID:           GenerateID(),
		ActivityID:   activityID,
		Filename:     "test_" + GenerateID()[:8] + ".pdf",
		OriginalName: "test.pdf",
		FilePath:     "/uploads/test.pdf",
		FileSize:     1024,
		ContentType:  "application/pdf",
		UploadedBy:   "",
		CreatedAt:    time.Now(),
	}

	// Apply overrides
	if overrides != nil {
		if v, ok := overrides["id"].(string); ok {
			attachment.ID = v
		}
		if v, ok := overrides["filename"].(string); ok {
			attachment.Filename = v
		}
		if v, ok := overrides["original_name"].(string); ok {
			attachment.OriginalName = v
		}
		if v, ok := overrides["file_path"].(string); ok {
			attachment.FilePath = v
		}
		if v, ok := overrides["file_size"].(int64); ok {
			attachment.FileSize = v
		}
		if v, ok := overrides["content_type"].(string); ok {
			attachment.ContentType = v
		}
		if v, ok := overrides["uploaded_by"].(string); ok {
			attachment.UploadedBy = v
		}
	}

	if err := fb.db.Create(attachment).Error; err != nil {
		return nil, err
	}

	return attachment, nil
}

// CreateMultipleUsers creates multiple users at once
func (fb *FixtureBuilder) CreateMultipleUsers(count int, overridesTemplate map[string]interface{}) ([]*User, error) {
	users := make([]*User, count)
	for i := 0; i < count; i++ {
		user, err := fb.CreateUser(overridesTemplate)
		if err != nil {
			return nil, err
		}
		users[i] = user
	}
	return users, nil
}

// CreateMultipleActivities creates multiple activities at once
func (fb *FixtureBuilder) CreateMultipleActivities(count int, overridesTemplate map[string]interface{}) ([]*Activity, error) {
	activities := make([]*Activity, count)
	for i := 0; i < count; i++ {
		activity, err := fb.CreateActivity(overridesTemplate)
		if err != nil {
			return nil, err
		}
		activities[i] = activity
	}
	return activities, nil
}

// SetupBasicScenario creates a basic test scenario with users and activities
func (fb *FixtureBuilder) SetupBasicScenario() (map[string]interface{}, error) {
	// Create departments
	csDept, err := fb.CreateDepartment(map[string]interface{}{
		"name": "Computer Science",
		"code": "CS",
	})
	if err != nil {
		return nil, err
	}

	mathDept, err := fb.CreateDepartment(map[string]interface{}{
		"name": "Mathematics",
		"code": "MATH",
	})
	if err != nil {
		return nil, err
	}

	// Create organizer
	organizer, err := fb.CreateUser(map[string]interface{}{
		"username":   "organizer",
		"name":       "Activity Organizer",
		"email":      "organizer@example.com",
		"department": csDept.Name,
		"role":       "teacher",
	})
	if err != nil {
		return nil, err
	}

	// Create students
	student1, err := fb.CreateUser(map[string]interface{}{
		"username":   "student1",
		"name":       "Student One",
		"email":      "student1@example.com",
		"department": csDept.Name,
		"role":       "student",
	})
	if err != nil {
		return nil, err
	}

	student2, err := fb.CreateUser(map[string]interface{}{
		"username":   "student2",
		"name":       "Student Two",
		"email":      "student2@example.com",
		"department": mathDept.Name,
		"role":       "student",
	})
	if err != nil {
		return nil, err
	}

	// Create admin
	admin, err := fb.CreateUser(map[string]interface{}{
		"username":   "admin",
		"name":       "System Admin",
		"email":      "admin@example.com",
		"department": "Administration",
		"role":       "admin",
	})
	if err != nil {
		return nil, err
	}

	// Create activities
	now := time.Now()
	activity1, err := fb.CreateActivity(map[string]interface{}{
		"name":               "Workshop: Go Programming",
		"type":               "workshop",
		"organizer_id":       organizer.ID,
		"start_time":         now.Add(24 * time.Hour),
		"end_time":           now.Add(26 * time.Hour),
		"registration_start": now,
		"registration_end":   now.Add(12 * time.Hour),
		"status":             "published",
		"base_credit":        2.0,
		"max_participants":   30,
	})
	if err != nil {
		return nil, err
	}

	activity2, err := fb.CreateActivity(map[string]interface{}{
		"name":               "Seminar: AI and Ethics",
		"type":               "seminar",
		"organizer_id":       organizer.ID,
		"start_time":         now.Add(48 * time.Hour),
		"end_time":           now.Add(50 * time.Hour),
		"registration_start": now,
		"registration_end":   now.Add(36 * time.Hour),
		"status":             "published",
		"base_credit":        1.5,
		"max_participants":   50,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"organizer":  organizer,
		"student1":   student1,
		"student2":   student2,
		"admin":      admin,
		"activity1":  activity1,
		"activity2":  activity2,
		"cs_dept":    csDept,
		"math_dept":  mathDept,
	}, nil
}
