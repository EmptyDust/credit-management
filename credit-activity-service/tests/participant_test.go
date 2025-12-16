package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"credit-management/credit-activity-service/models"
	testutils "credit-management/test-utils"
)

// TestAddParticipant tests adding a participant to an activity
func TestAddParticipant(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "activity_participants")
	ah := testutils.NewAssertHelper(t)

	// Create approved activity
	activity := models.CreditActivity{
		Title:       "Workshop with Participants",
		Description: "Test activity",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Status:      models.StatusApproved,
		Category:    models.CategoryInnovation,
		OwnerID:     testutils.GenerateID(),
	}
	err := testDB.DB.Create(&activity).Error
	ah.RequireNoError(err)

	// Create participant request
	userID := testutils.GenerateID()
	participantReq := map[string]interface{}{
		"user_id": userID,
		"credits": 2.5,
	}

	// Setup router
	router := gin.New()
	router.POST("/api/activities/:id/participants", func(c *gin.Context) {
		c.Set("id", testutils.GenerateID())
		c.Set("user_type", "admin")
		activityHandler.AddParticipant(c)
	})

	req, err := testutils.CreateJSONRequest("POST", "/api/activities/"+activity.ID+"/participants", participantReq)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	ah.AssertHTTPStatus(resp, http.StatusCreated)
	ah.AssertJSONFieldExists(resp, "data.id")
	ah.AssertJSONFieldEquals(resp, "data.activity_id", activity.ID)

	// Verify participant was added to database
	var participant models.ActivityParticipant
	err = testDB.DB.Where("activity_id = ? AND user_id = ?", activity.ID, userID).First(&participant).Error
	ah.RequireNoError(err)
	ah.AssertEqual(2.5, participant.Credits)
}

// TestGetActivityParticipants tests listing participants of an activity
func TestGetActivityParticipants(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "activity_participants")
	ah := testutils.NewAssertHelper(t)

	// Create activity
	activity := models.CreditActivity{
		Title:       "Activity with Multiple Participants",
		Description: "Test activity",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Status:      models.StatusApproved,
		Category:    models.CategoryInnovation,
		OwnerID:     testutils.GenerateID(),
	}
	err := testDB.DB.Create(&activity).Error
	ah.RequireNoError(err)

	// Add multiple participants
	for i := 0; i < 5; i++ {
		participant := models.ActivityParticipant{
			ActivityID: activity.ID,
			UUID:       testutils.GenerateID(),
			Credits:    2.0,
			JoinedAt:   time.Now(),
		}
		err := testDB.DB.Create(&participant).Error
		ah.RequireNoError(err)
	}

	// Setup router
	router := gin.New()
	router.GET("/api/activities/:id/participants", func(c *gin.Context) {
		c.Set("id", testutils.GenerateID())
		c.Set("user_type", "student")
		activityHandler.GetActivityParticipants(c)
	})

	req, err := testutils.CreateJSONRequest("GET", "/api/activities/"+activity.ID+"/participants", nil)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	ah.AssertHTTPStatus(resp, http.StatusOK)
	ah.AssertJSONArrayLength(resp, "data", 5)
}

// TestRemoveParticipant tests removing a participant from an activity
func TestRemoveParticipant(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "activity_participants")
	ah := testutils.NewAssertHelper(t)

	// Create activity
	activity := models.CreditActivity{
		Title:       "Activity for Removal Test",
		Description: "Test activity",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Status:      models.StatusApproved,
		Category:    models.CategoryInnovation,
		OwnerID:     testutils.GenerateID(),
	}
	err := testDB.DB.Create(&activity).Error
	ah.RequireNoError(err)

	// Add participant
	participant := models.ActivityParticipant{
		ActivityID: activity.ID,
		UUID:       testutils.GenerateID(),
		Credits:    2.0,
		JoinedAt:   time.Now(),
	}
	err = testDB.DB.Create(&participant).Error
	ah.RequireNoError(err)

	// Setup router
	router := gin.New()
	router.DELETE("/api/activities/:activity_id/participants/:participant_id", func(c *gin.Context) {
		c.Set("id", testutils.GenerateID())
		c.Set("user_type", "admin")
		activityHandler.RemoveParticipant(c)
	})

	req, err := testutils.CreateJSONRequest("DELETE", "/api/activities/"+activity.ID+"/participants/"+participant.ID, nil)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	ah.AssertHTTPStatus(resp, http.StatusOK)

	// Verify participant was removed (soft delete)
	var deletedParticipant models.ActivityParticipant
	err = testDB.DB.Unscoped().Where("id = ?", participant.ID).First(&deletedParticipant).Error
	ah.RequireNoError(err)
	ah.AssertNotNil(deletedParticipant.DeletedAt)
}

// TestUpdateParticipantCredits tests updating credits for a participant
func TestUpdateParticipantCredits(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "activity_participants")
	ah := testutils.NewAssertHelper(t)

	// Create activity
	activity := models.CreditActivity{
		Title:       "Activity for Credit Update",
		Description: "Test activity",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Status:      models.StatusApproved,
		Category:    models.CategoryInnovation,
		OwnerID:     testutils.GenerateID(),
	}
	err := testDB.DB.Create(&activity).Error
	ah.RequireNoError(err)

	// Add participant
	participant := models.ActivityParticipant{
		ActivityID: activity.ID,
		UUID:       testutils.GenerateID(),
		Credits:    2.0,
		JoinedAt:   time.Now(),
	}
	err = testDB.DB.Create(&participant).Error
	ah.RequireNoError(err)

	// Setup router
	router := gin.New()
	router.PUT("/api/activities/:activity_id/participants/:participant_id", func(c *gin.Context) {
		c.Set("id", testutils.GenerateID())
		c.Set("user_type", "admin")
		activityHandler.UpdateParticipant(c)
	})

	updateReq := map[string]interface{}{
		"credits": 3.5,
	}

	req, err := testutils.CreateJSONRequest("PUT", "/api/activities/"+activity.ID+"/participants/"+participant.ID, updateReq)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	ah.AssertHTTPStatus(resp, http.StatusOK)

	// Verify credits were updated
	var updatedParticipant models.ActivityParticipant
	err = testDB.DB.Where("id = ?", participant.ID).First(&updatedParticipant).Error
	ah.RequireNoError(err)
	ah.AssertEqual(3.5, updatedParticipant.Credits)
}

// TestDuplicateParticipant tests preventing duplicate participants
func TestDuplicateParticipant(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "activity_participants")
	ah := testutils.NewAssertHelper(t)

	// Create activity
	activity := models.CreditActivity{
		Title:       "Activity for Duplicate Test",
		Description: "Test activity",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Status:      models.StatusApproved,
		Category:    models.CategoryInnovation,
		OwnerID:     testutils.GenerateID(),
	}
	err := testDB.DB.Create(&activity).Error
	ah.RequireNoError(err)

	userID := testutils.GenerateID()

	// Add participant first time
	participant := models.ActivityParticipant{
		ActivityID: activity.ID,
		UUID:       userID,
		Credits:    2.0,
		JoinedAt:   time.Now(),
	}
	err = testDB.DB.Create(&participant).Error
	ah.RequireNoError(err)

	// Try to add same participant again
	router := gin.New()
	router.POST("/api/activities/:id/participants", func(c *gin.Context) {
		c.Set("id", testutils.GenerateID())
		c.Set("user_type", "admin")
		activityHandler.AddParticipant(c)
	})

	participantReq := map[string]interface{}{
		"user_id": userID,
		"credits": 2.5,
	}

	req, err := testutils.CreateJSONRequest("POST", "/api/activities/"+activity.ID+"/participants", participantReq)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	// Should return error for duplicate
	ah.AssertHTTPStatus(resp, http.StatusBadRequest)
}

// TestBatchAddParticipants tests adding multiple participants at once
func TestBatchAddParticipants(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "activity_participants")
	ah := testutils.NewAssertHelper(t)

	// Create activity
	activity := models.CreditActivity{
		Title:       "Activity for Batch Add",
		Description: "Test activity",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Status:      models.StatusApproved,
		Category:    models.CategoryInnovation,
		OwnerID:     testutils.GenerateID(),
	}
	err := testDB.DB.Create(&activity).Error
	ah.RequireNoError(err)

	// Setup router
	router := gin.New()
	router.POST("/api/activities/:id/participants/batch", func(c *gin.Context) {
		c.Set("id", testutils.GenerateID())
		c.Set("user_type", "admin")
		activityHandler.BatchAddParticipants(c)
	})

	// Batch add request
	batchReq := map[string]interface{}{
		"participants": []map[string]interface{}{
			{"user_id": testutils.GenerateID(), "credits": 2.0},
			{"user_id": testutils.GenerateID(), "credits": 2.5},
			{"user_id": testutils.GenerateID(), "credits": 3.0},
		},
	}

	req, err := testutils.CreateJSONRequest("POST", "/api/activities/"+activity.ID+"/participants/batch", batchReq)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	ah.AssertHTTPStatus(resp, http.StatusCreated)

	// Verify all participants were added
	var count int64
	testDB.DB.Model(&models.ActivityParticipant{}).Where("activity_id = ?", activity.ID).Count(&count)
	ah.AssertEqual(int64(3), count)
}

// TestParticipantPermissions tests permission checks for participant management
func TestParticipantPermissions(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "activity_participants")
	ah := testutils.NewAssertHelper(t)

	// Create activity
	activity := models.CreditActivity{
		Title:       "Activity for Permission Test",
		Description: "Test activity",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Status:      models.StatusApproved,
		Category:    models.CategoryInnovation,
		OwnerID:     testutils.GenerateID(),
	}
	err := testDB.DB.Create(&activity).Error
	ah.RequireNoError(err)

	// Try to add participant as student (should fail)
	router := gin.New()
	router.POST("/api/activities/:id/participants", func(c *gin.Context) {
		c.Set("id", testutils.GenerateID())
		c.Set("user_type", "student") // Not admin
		activityHandler.AddParticipant(c)
	})

	participantReq := map[string]interface{}{
		"user_id": testutils.GenerateID(),
		"credits": 2.0,
	}

	req, err := testutils.CreateJSONRequest("POST", "/api/activities/"+activity.ID+"/participants", participantReq)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	// Should be forbidden
	assert.True(t, resp.Code == http.StatusForbidden || resp.Code == http.StatusUnauthorized)
}

// TestGetParticipantStatistics tests getting statistics for participants
func TestGetParticipantStatistics(t *testing.T) {
	testDB.CleanDatabase("credit_activities", "activity_participants")
	ah := testutils.NewAssertHelper(t)

	// Create multiple activities
	userID := testutils.GenerateID()

	for i := 0; i < 3; i++ {
		activity := models.CreditActivity{
			Title:       "Activity " + string(rune(i+1)),
			Description: "Test activity",
			StartDate:   time.Now().Add(24 * time.Hour),
			EndDate:     time.Now().Add(48 * time.Hour),
			Status:      models.StatusApproved,
			Category:    models.CategoryInnovation,
			OwnerID:     testutils.GenerateID(),
		}
		err := testDB.DB.Create(&activity).Error
		ah.RequireNoError(err)

		// Add same user as participant with different credits
		participant := models.ActivityParticipant{
			ActivityID: activity.ID,
			UUID:       userID,
			Credits:    float64(i + 1),
			JoinedAt:   time.Now(),
		}
		err = testDB.DB.Create(&participant).Error
		ah.RequireNoError(err)
	}

	// Get participant statistics
	router := gin.New()
	router.GET("/api/participants/:user_id/stats", func(c *gin.Context) {
		c.Set("id", userID)
		c.Set("user_type", "student")
		activityHandler.GetParticipantStats(c)
	})

	req, err := testutils.CreateJSONRequest("GET", "/api/participants/"+userID+"/stats", nil)
	ah.RequireNoError(err)

	resp := testutils.PerformRequest(router, req)

	ah.AssertHTTPStatus(resp, http.StatusOK)
	ah.AssertJSONFieldExists(resp, "data.total_activities")
	ah.AssertJSONFieldExists(resp, "data.total_credits")
}
