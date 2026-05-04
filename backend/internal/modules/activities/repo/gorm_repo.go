package repo

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	activitydomain "credit-management/backend/internal/modules/activities/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const emptyDateSentinel = "1970-01-01"

type GormRepository struct {
	db *gorm.DB
}

type activityRow struct {
	ID             string     `gorm:"column:id"`
	Title          string     `gorm:"column:title"`
	Description    *string    `gorm:"column:description"`
	CategoryID     string     `gorm:"column:category_id"`
	CategoryKey    string     `gorm:"column:category_key;->"`
	AcademicYear   string     `gorm:"column:academic_year"`
	Semester       string     `gorm:"column:semester"`
	StartDate      time.Time  `gorm:"column:start_date"`
	EndDate        time.Time  `gorm:"column:end_date"`
	Status         string     `gorm:"column:status"`
	OwnerID        string     `gorm:"column:owner_id"`
	ReviewerID     *string    `gorm:"column:reviewer_id"`
	ReviewComments *string    `gorm:"column:review_comments"`
	ReviewedAt     *time.Time `gorm:"column:reviewed_at"`
	DetailsJSON    []byte     `gorm:"column:details_json"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

type categoryRow struct {
	ID    string  `gorm:"column:id"`
	Key   string  `gorm:"column:key"`
	Label string  `gorm:"column:label"`
	Sort  int     `gorm:"column:sort_order"`
	On    bool    `gorm:"column:enabled"`
	Del   *string `gorm:"column:deleted_at"`
}

type activityCategoryBindingRow struct {
	CategoryID  string `gorm:"column:category_id"`
	CategoryKey string `gorm:"column:category_key"`
}

func (activityRow) TableName() string {
	return "activities"
}

func (categoryRow) TableName() string {
	return "activity_category_templates"
}

func NewGorm(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) List(ctx context.Context, filter ListFilter) ([]activitydomain.Activity, error) {
	var rows []activityRow

	query := r.baseQuery(ctx)
	if filter.OwnerID != "" {
		query = query.Where("activities.owner_id = ?", filter.OwnerID)
	}
	if filter.Status != "" {
		query = query.Where("activities.status = ?", filter.Status)
	}
	if filter.Category != "" {
		query = query.Where("activity_category_templates.key = ?", filter.Category)
	}
	if filter.AcademicYear != "" {
		query = query.Where("activities.academic_year = ?", filter.AcademicYear)
	}
	if filter.Semester != "" {
		query = query.Where("activities.semester = ?", filter.Semester)
	}
	if filter.Keyword != "" {
		query = query.Where("activities.title ILIKE ?", "%"+strings.TrimSpace(filter.Keyword)+"%")
	}

	if err := query.Order("activities.created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]activitydomain.Activity, 0, len(rows))
	for i := range rows {
		items = append(items, rows[i].toDomain())
	}
	return items, nil
}

func (r *GormRepository) Get(ctx context.Context, activityID string) (*activitydomain.Activity, error) {
	var row activityRow
	if err := r.baseQuery(ctx).First(&row, "activities.id = ?", activityID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, activitydomain.ErrActivityNotFound
		}
		return nil, err
	}

	activity := row.toDomain()
	return &activity, nil
}

func (r *GormRepository) Save(ctx context.Context, activity *activitydomain.Activity) error {
	categoryID, err := r.resolveCategoryID(ctx, activity.ID, activity.Category)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(cloneDetails(activity.Details))
	if err != nil {
		return err
	}

	row := activityRow{
		ID:             activity.ID,
		Title:          activity.Title,
		Description:    nullableString(activity.Description),
		CategoryID:     categoryID,
		AcademicYear:   activity.AcademicYear,
		Semester:       activity.Semester,
		StartDate:      parseDateOrSentinel(activity.StartDate),
		EndDate:        parseDateOrSentinel(activity.EndDate),
		Status:         activity.Status,
		OwnerID:        activity.OwnerID,
		ReviewerID:     nullableString(activity.ReviewerID),
		ReviewComments: nullableString(activity.ReviewComments),
		ReviewedAt:     cloneTimePtr(activity.ReviewedAt),
		DetailsJSON:    payload,
		DeletedAt:      cloneTimePtr(activity.DeletedAt),
	}

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "description", "category_id", "academic_year", "semester", "start_date", "end_date", "status", "owner_id", "reviewer_id", "review_comments", "reviewed_at", "details_json", "deleted_at", "updated_at"}),
	}).Create(&row).Error
}

func (r *GormRepository) Delete(ctx context.Context, activityID string) error {
	result := r.db.WithContext(ctx).
		Model(&activityRow{}).
		Where("id = ? AND deleted_at IS NULL", activityID).
		Update("deleted_at", time.Now().UTC())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return activitydomain.ErrActivityNotFound
	}
	return nil
}

func (r *GormRepository) baseQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("activities").
		Select("activities.*, activity_category_templates.key AS category_key").
		Joins("JOIN activity_category_templates ON activity_category_templates.id = activities.category_id").
		Where("activities.deleted_at IS NULL")
}

func (r *GormRepository) resolveCategoryID(ctx context.Context, activityID string, key string) (string, error) {
	key = normalizeCategoryKey(key)

	if activityID != "" {
		binding, found, err := r.findExistingCategoryBinding(ctx, activityID)
		if err != nil {
			return "", err
		}
		if found && binding.CategoryKey == key {
			return binding.CategoryID, nil
		}
	}

	var row categoryRow
	if err := r.db.WithContext(ctx).
		Where("key = ? AND enabled = ? AND deleted_at IS NULL", key, true).
		First(&row).Error; err == nil {
		return row.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	return "", ErrCategorySchemaNotFound
}

func (r *GormRepository) findExistingCategoryBinding(ctx context.Context, activityID string) (activityCategoryBindingRow, bool, error) {
	var row activityCategoryBindingRow
	err := r.db.WithContext(ctx).
		Table("activities").
		Select("activities.category_id", "activity_category_templates.key AS category_key").
		Joins("JOIN activity_category_templates ON activity_category_templates.id = activities.category_id").
		Where("activities.id = ? AND activities.deleted_at IS NULL AND activity_category_templates.deleted_at IS NULL", activityID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return activityCategoryBindingRow{}, false, nil
	}
	if err != nil {
		return activityCategoryBindingRow{}, false, err
	}
	return row, true, nil
}

func normalizeCategoryKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "general"
	}
	return key
}

func (r activityRow) toDomain() activitydomain.Activity {
	var details map[string]any
	if len(r.DetailsJSON) > 0 {
		_ = json.Unmarshal(r.DetailsJSON, &details)
	}
	if details == nil {
		details = map[string]any{}
	}

	return activitydomain.Activity{
		ID:             r.ID,
		Title:          r.Title,
		Description:    derefString(r.Description),
		Category:       r.CategoryKey,
		AcademicYear:   r.AcademicYear,
		Semester:       r.Semester,
		StartDate:      formatDateOrEmpty(r.StartDate),
		EndDate:        formatDateOrEmpty(r.EndDate),
		Status:         r.Status,
		OwnerID:        r.OwnerID,
		ReviewerID:     derefString(r.ReviewerID),
		ReviewComments: derefString(r.ReviewComments),
		ReviewedAt:     cloneTimePtr(r.ReviewedAt),
		Details:        details,
		DeletedAt:      cloneTimePtr(r.DeletedAt),
	}
}

func parseDateOrSentinel(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		value = emptyDateSentinel
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		fallback, _ := time.Parse("2006-01-02", emptyDateSentinel)
		return fallback
	}
	return parsed
}

func formatDateOrEmpty(value time.Time) string {
	text := value.Format("2006-01-02")
	if text == emptyDateSentinel {
		return ""
	}
	return text
}

func cloneDetails(details map[string]any) map[string]any {
	if details == nil {
		return map[string]any{}
	}
	copy := make(map[string]any, len(details))
	for key, value := range details {
		copy[key] = value
	}
	return copy
}

func nullableString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	t := value.UTC()
	return &t
}
