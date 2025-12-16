package tests

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"credit-management/credit-activity-service/handlers"
	"credit-management/credit-activity-service/models"
	"credit-management/credit-activity-service/utils"
	testutils "credit-management/test-utils"
)

var (
	testDB          *testutils.TestDatabase
	testRouter      *gin.Engine
	activityHandler *handlers.ActivityHandler
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Set up test database
	var err error
	testDB, err = testutils.SetupTestDatabase(ctx)
	if err != nil {
		panic("Failed to setup test database: " + err.Error())
	}

	// Auto-migrate models
	err = testDB.DB.AutoMigrate(
		&models.CreditActivity{},
		&models.ActivityParticipant{},
		&models.Application{},
		&models.Attachment{},
	)
	if err != nil {
		panic("Failed to migrate models: " + err.Error())
	}

	// Initialize handlers
	validator := utils.NewValidator()
	activityHandler = handlers.NewActivityHandler(testDB.DB, validator)

	// Set up Gin router
	gin.SetMode(gin.TestMode)
	testRouter = gin.New()

	// Register routes
	api := testRouter.Group("/api")
	{
		activities := api.Group("/activities")
		{
			activities.POST("", mockAuthMiddleware("student"), activityHandler.CreateActivity)
			activities.GET("", mockAuthMiddleware("student"), activityHandler.GetActivities)
			activities.GET("/:id", mockAuthMiddleware("student"), activityHandler.GetActivity)
			activities.PUT("/:id", mockAuthMiddleware("student"), activityHandler.UpdateActivity)
			activities.DELETE("/:id", mockAuthMiddleware("student"), activityHandler.DeleteActivity)
			activities.POST("/:id/submit", mockAuthMiddleware("student"), activityHandler.SubmitForReview)
			activities.POST("/:id/approve", mockAuthMiddleware("admin"), activityHandler.ApproveActivity)
			activities.POST("/:id/reject", mockAuthMiddleware("admin"), activityHandler.RejectActivity)
		}
	}

	// Run tests
	code := m.Run()

	// Cleanup
	testDB.Teardown(ctx)

	os.Exit(code)
}

// mockAuthMiddleware simulates authentication middleware for testing
func mockAuthMiddleware(userType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("id", testutils.GenerateID())
		c.Set("user_type", userType)
		c.Set("username", "testuser")
		c.Next()
	}
}

// TestCreateActivity tests creating a new activity
func TestCreateActivity(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "activity_participants", "applications", "attachments")
	ah := testutils.NewAssertHelper(t)

	activityReq := models.ActivityRequest{
		Title:       "Go Programming Workshop",
		Description: "Learn Go programming from basics to advanced",
		StartDate:   "2024-12-20",
		EndDate:     "2024-12-22",
		Category:    models.CategoryInnovation,
		Details: map[string]interface{}{
			"location":     "Room 101",
			"max_capacity": 50,
		},
	}

	req, err := testutils.CreateJSONRequest("POST", "/api/activities", activityReq)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(testRouter, req)

	ah.AssertHTTPStatus(resp, http.StatusCreated)
	ah.AssertJSONFieldExists(resp, "data.id")
	ah.AssertJSONFieldEquals(resp, "data.title", "Go Programming Workshop")
	ah.AssertJSONFieldEquals(resp, "data.status", models.StatusDraft)
	ah.AssertJSONFieldEquals(resp, "data.category", models.CategoryInnovation)
}

// TestCreateActivityValidation tests activity creation validation
func TestCreateActivityValidation(t *testing.T) {
	testDB.CleanDatabase("credit_activities")
	ah := testutils.NewAssertHelper(t)

	tests := []struct {
		name        string
		request     models.ActivityRequest
		expectError string
	}{
		{
			name: "missing title",
			request: models.ActivityRequest{
				Description: "Test description",
				StartDate:   "2024-12-20",
				EndDate:     "2024-12-22",
				Category:    models.CategoryInnovation,
			},
			expectError: "标题不能为空",
		},
		{
			name: "missing category",
			request: models.ActivityRequest{
				Title:       "Test Activity",
				Description: "Test description",
				StartDate:   "2024-12-20",
				EndDate:     "2024-12-22",
			},
			expectError: "类别不能为空",
		},
		{
			name: "invalid date range",
			request: models.ActivityRequest{
				Title:       "Test Activity",
				Description: "Test description",
				StartDate:   "2024-12-22",
				EndDate:     "2024-12-20",
				Category:    models.CategoryInnovation,
			},
			expectError: "开始日期不能晚于结束日期",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := testutils.CreateJSONRequest("POST", "/api/activities", tt.request)
			require.NoError(t, err)

			resp := testutils.PerformRequest(testRouter, req)

			ah.AssertHTTPStatus(resp, http.StatusBadRequest)
			ah.AssertErrorResponse(resp, http.StatusBadRequest, tt.expectError)
		})
	}
}

// TestGetActivities tests listing activities
func TestGetActivities(t *testing.T) {
	testDB.CleanDatabase("credit_activities")
	ah := testutils.NewAssertHelper(t)

	// Create test activities
	userID := testutils.GenerateID()
	activities := []models.CreditActivity{
		{
			Title:       "Activity 1",
			Description: "Description 1",
			StartDate:   time.Now().Add(24 * time.Hour),
			EndDate:     time.Now().Add(48 * time.Hour),
			Status:      models.StatusDraft,
			Category:    models.CategoryInnovation,
			OwnerID:     userID,
		},
		{
			Title:       "Activity 2",
			Description: "Description 2",
			StartDate:   time.Now().Add(72 * time.Hour),
			EndDate:     time.Now().Add(96 * time.Hour),
			Status:      models.StatusApproved,
			Category:    models.CategoryCompetition,
			OwnerID:     userID,
		},
	}

	for _, activity := range activities {
		err := testDB.DB.Create(&activity).Error
		ah.RequireNoError(err)
	}

	req, err := testutils.CreateJSONRequest("GET", "/api/activities?page=1&page_size=10", nil)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(testRouter, req)

	ah.AssertHTTPStatus(resp, http.StatusOK)
	ah.AssertJSONArrayLength(resp, "data", 2)
	ah.AssertPaginationResponse(resp, 2)
}

// TestGetActivityByID tests getting a specific activity
func TestGetActivityByID(t *testing.T) {
	testDB.CleanDatabase("credit_activities")
	ah := testutils.NewAssertHelper(t)

	// Create test activity
	activity := models.CreditActivity{
		Title:       "Test Activity",
		Description: "Test Description",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Status:      models.StatusDraft,
		Category:    models.CategoryInnovation,
		OwnerID:     testutils.GenerateID(),
	}

	err := testDB.DB.Create(&activity).Error
	ah.RequireNoError(err)

	req, err := testutils.CreateJSONRequest("GET", "/api/activities/"+activity.ID, nil)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(testRouter, req)

	ah.AssertHTTPStatus(resp, http.StatusOK)
	ah.AssertJSONFieldEquals(resp, "data.id", activity.ID)
	ah.AssertJSONFieldEquals(resp, "data.title", "Test Activity")
}

// TestUpdateActivity tests updating an activity
func TestUpdateActivity(t *testing.T) {
	testDB.CleanDatabase("credit_activities")
	ah := testutils.NewAssertHelper(t)

	// Create test activity
	userID := testutils.GenerateID()
	activity := models.CreditActivity{
		Title:       "Original Title",
		Description: "Original Description",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Status:      models.StatusDraft,
		Category:    models.CategoryInnovation,
		OwnerID:     userID,
	}

	err := testDB.DB.Create(&activity).Error
	ah.RequireNoError(err)

	// Update request
	updateReq := models.ActivityRequest{
		Title:       "Updated Title",
		Description: "Updated Description",
		StartDate:   "2024-12-25",
		EndDate:     "2024-12-27",
		Category:    models.CategoryCompetition,
	}

	// Create custom router with user context
	router := gin.New()
	router.PUT("/api/activities/:id", func(c *gin.Context) {
		c.Set("id", userID)
		c.Set("user_type", "student")
		activityHandler.UpdateActivity(c)
	})

	req, err := testutils.CreateJSONRequest("PUT", "/api/activities/"+activity.ID, updateReq)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	ah.AssertHTTPStatus(resp, http.StatusOK)
	ah.AssertJSONFieldEquals(resp, "data.title", "Updated Title")
	ah.AssertJSONFieldEquals(resp, "data.category", models.CategoryCompetition)
}

// TestDeleteActivity tests deleting an activity
func TestDeleteActivity(t *testing.T) {
	testDB.CleanDatabase("credit_activities")
	ah := testutils.NewAssertHelper(t)

	// Create test activity
	userID := testutils.GenerateID()
	activity := models.CreditActivity{
		Title:       "Activity to Delete",
		Description: "This will be deleted",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Status:      models.StatusDraft,
		Category:    models.CategoryInnovation,
		OwnerID:     userID,
	}

	err := testDB.DB.Create(&activity).Error
	ah.RequireNoError(err)

	// Create custom router with user context
	router := gin.New()
	router.DELETE("/api/activities/:id", func(c *gin.Context) {
		c.Set("id", userID)
		c.Set("user_type", "student")
		activityHandler.DeleteActivity(c)
	})

	req, err := testutils.CreateJSONRequest("DELETE", "/api/activities/"+activity.ID, nil)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	ah.AssertHTTPStatus(resp, http.StatusOK)

	// Verify activity is soft-deleted
	var deletedActivity models.CreditActivity
	err = testDB.DB.Unscoped().First(&deletedActivity, "id = ?", activity.ID).Error
	ah.RequireNoError(err)
	ah.AssertNotNil(deletedActivity.DeletedAt)
}

// TestActivityLifecycle tests the complete activity lifecycle
func TestActivityLifecycle(t *testing.T) {
	testDB.CleanDatabase("credit_activities")
	ah := testutils.NewAssertHelper(t)

	// 1. Create activity (draft)
	userID := testutils.GenerateID()
	activity := models.CreditActivity{
		Title:       "Lifecycle Test Activity",
		Description: "Testing activity lifecycle",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Status:      models.StatusDraft,
		Category:    models.CategoryInnovation,
		OwnerID:     userID,
	}

	err := testDB.DB.Create(&activity).Error
	ah.RequireNoError(err)
	ah.AssertEqual(models.StatusDraft, activity.Status)

	// 2. Submit for review
	router := gin.New()
	router.POST("/api/activities/:id/submit", func(c *gin.Context) {
		c.Set("id", userID)
		c.Set("user_type", "student")
		activityHandler.SubmitForReview(c)
	})

	req, err := testutils.CreateJSONRequest("POST", "/api/activities/"+activity.ID+"/submit", nil)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)
	ah.AssertSuccessResponse(resp)

	// Verify status changed to pending_review
	var updatedActivity models.CreditActivity
	testDB.DB.First(&updatedActivity, "id = ?", activity.ID)
	ah.AssertEqual(models.StatusPendingReview, updatedActivity.Status)

	// 3. Approve activity (as admin)
	adminRouter := gin.New()
	adminRouter.POST("/api/activities/:id/approve", func(c *gin.Context) {
		c.Set("id", testutils.GenerateID())
		c.Set("user_type", "admin")
		activityHandler.ApproveActivity(c)
	})

	approveReq := map[string]interface{}{
		"review_comments": "Approved! Great activity.",
	}

	req, err = testutils.CreateJSONRequest("POST", "/api/activities/"+activity.ID+"/approve", approveReq)
	ah.RequireNoError(err)

	resp = testutils.PerformRequest(adminRouter, req)
	ah.AssertSuccessResponse(resp)

	// Verify status changed to approved
	testDB.DB.First(&updatedActivity, "id = ?", activity.ID)
	ah.AssertEqual(models.StatusApproved, updatedActivity.Status)
	ah.AssertNotNil(updatedActivity.ReviewedAt)
	ah.AssertNotNil(updatedActivity.ReviewerID)
}

// TestActivityRejection tests rejecting an activity
func TestActivityRejection(t *testing.T) {
	testDB.CleanDatabase("credit_activities")
	ah := testutils.NewAssertHelper(t)

	// Create activity in pending_review status
	activity := models.CreditActivity{
		Title:       "Activity to Reject",
		Description: "This will be rejected",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Status:      models.StatusPendingReview,
		Category:    models.CategoryInnovation,
		OwnerID:     testutils.GenerateID(),
	}

	err := testDB.DB.Create(&activity).Error
	ah.RequireNoError(err)

	// Reject activity
	router := gin.New()
	router.POST("/api/activities/:id/reject", func(c *gin.Context) {
		c.Set("id", testutils.GenerateID())
		c.Set("user_type", "admin")
		activityHandler.RejectActivity(c)
	})

	rejectReq := map[string]interface{}{
		"review_comments": "Insufficient details provided.",
	}

	req, err := testutils.CreateJSONRequest("POST", "/api/activities/"+activity.ID+"/reject", rejectReq)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)
	ah.AssertSuccessResponse(resp)

	// Verify status changed to rejected
	var updatedActivity models.CreditActivity
	testDB.DB.First(&updatedActivity, "id = ?", activity.ID)
	ah.AssertEqual(models.StatusRejected, updatedActivity.Status)
	ah.AssertNotNil(updatedActivity.ReviewedAt)
	ah.AssertEqual("Insufficient details provided.", updatedActivity.ReviewComments)
}

// TestActivityFiltering tests filtering activities by various criteria
func TestActivityFiltering(t *testing.T) {
	testDB.CleanDatabase("credit_activities")
	ah := testutils.NewAssertHelper(t)

	// Create activities with different statuses and categories
	activities := []models.CreditActivity{
		{
			Title:     "Draft Innovation",
			Status:    models.StatusDraft,
			Category:  models.CategoryInnovation,
			OwnerID:   testutils.GenerateID(),
			StartDate: time.Now(),
			EndDate:   time.Now().Add(24 * time.Hour),
		},
		{
			Title:     "Approved Competition",
			Status:    models.StatusApproved,
			Category:  models.CategoryCompetition,
			OwnerID:   testutils.GenerateID(),
			StartDate: time.Now(),
			EndDate:   time.Now().Add(24 * time.Hour),
		},
		{
			Title:     "Pending Innovation",
			Status:    models.StatusPendingReview,
			Category:  models.CategoryInnovation,
			OwnerID:   testutils.GenerateID(),
			StartDate: time.Now(),
			EndDate:   time.Now().Add(24 * time.Hour),
		},
	}

	for _, activity := range activities {
		err := testDB.DB.Create(&activity).Error
		ah.RequireNoError(err)
	}

	// Test filtering by status
	req, _ := testutils.CreateJSONRequest("GET", "/api/activities?status="+models.StatusApproved, nil)
	resp := testutils.PerformRequest(testRouter, req)
	ah.AssertHTTPStatus(resp, http.StatusOK)
	// Should return 1 approved activity

	// Test filtering by category
	req, _ = testutils.CreateJSONRequest("GET", "/api/activities?category="+models.CategoryInnovation, nil)
	resp = testutils.PerformRequest(testRouter, req)
	ah.AssertHTTPStatus(resp, http.StatusOK)
	// Should return 2 innovation activities
}
