package repo

import (
	"context"

	optionsservice "credit-management/backend/internal/modules/options/service"

	"gorm.io/gorm"
)

type GormOptionRepo struct {
	db *gorm.DB
}

func NewGorm(db *gorm.DB) *GormOptionRepo {
	return &GormOptionRepo{db: db}
}

func (r *GormOptionRepo) ListOrganization(ctx context.Context) ([]optionsservice.OrganizationNode, error) {
	type organizationRow struct {
		ID        string
		ParentID  *string
		NodeType  string
		Name      string
		Path      string
		Level     int
		SortOrder int
	}

	rows := make([]organizationRow, 0)
	err := r.db.WithContext(ctx).
		Table("organization_nodes").
		Select("id", "parent_id", "node_type", "name", "path", "level", "sort_order").
		Where("deleted_at is null").
		Order("level asc, sort_order asc, path asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	nodes := make([]optionsservice.OrganizationNode, 0, len(rows))
	for _, row := range rows {
		parentID := ""
		if row.ParentID != nil {
			parentID = *row.ParentID
		}
		nodes = append(nodes, optionsservice.OrganizationNode{
			ID:        row.ID,
			ParentID:  parentID,
			NodeType:  row.NodeType,
			Name:      row.Name,
			Path:      row.Path,
			Level:     row.Level,
			SortOrder: row.SortOrder,
		})
	}

	return nodes, nil
}

func (r *GormOptionRepo) ListOptionItems(ctx context.Context, groupKey string) ([]optionsservice.OptionItem, error) {
	type optionItemRow struct {
		Value string
		Label string
	}

	rows := make([]optionItemRow, 0)
	err := r.db.WithContext(ctx).
		Table("option_items").
		Select("option_items.value", "option_items.label").
		Joins("join option_groups on option_groups.id = option_items.group_id").
		Where("option_groups.key = ? and option_groups.scope = ? and option_groups.deleted_at is null and option_items.deleted_at is null and option_items.enabled = ?", groupKey, "system", true).
		Order("option_items.sort_order asc, option_items.value asc").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]optionsservice.OptionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, optionsservice.OptionItem{
			Value: row.Value,
			Label: row.Label,
		})
	}

	return items, nil
}
