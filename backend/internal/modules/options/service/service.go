package service

import (
	"context"
	"sort"
)

type OrganizationNode struct {
	ID        string
	ParentID  string
	NodeType  string
	Name      string
	Path      string
	Level     int
	SortOrder int
}

type OptionItem struct {
	Value string
	Label string
}

type Repository interface {
	ListOrganization(ctx context.Context) ([]OrganizationNode, error)
	ListOptionItems(ctx context.Context, groupKey string) ([]OptionItem, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) OrganizationTree(ctx context.Context) ([]map[string]any, error) {
	nodes, err := s.repo.ListOrganization(ctx)
	if err != nil {
		return nil, err
	}

	return buildTree(nodes), nil
}

func (s *Service) OptionList(ctx context.Context, groupKey string) ([]map[string]any, error) {
	items, err := s.repo.ListOptionItems(ctx, groupKey)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"value": item.Value,
			"label": item.Label,
		})
	}

	return result, nil
}

func buildTree(nodes []OrganizationNode) []map[string]any {
	if len(nodes) == 0 {
		return []map[string]any{}
	}

	sorted := append([]OrganizationNode(nil), nodes...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Level != sorted[j].Level {
			return sorted[i].Level < sorted[j].Level
		}
		if sorted[i].SortOrder != sorted[j].SortOrder {
			return sorted[i].SortOrder < sorted[j].SortOrder
		}
		return sorted[i].Path < sorted[j].Path
	})

	index := make(map[string]map[string]any, len(sorted))
	roots := make([]map[string]any, 0)
	for _, node := range sorted {
		payload := map[string]any{
			"id":        node.ID,
			"name":      node.Name,
			"node_type": node.NodeType,
			"children":  []map[string]any{},
		}
		index[node.ID] = payload

		if node.ParentID == "" {
			roots = append(roots, payload)
			continue
		}

		parent, ok := index[node.ParentID]
		if !ok {
			roots = append(roots, payload)
			continue
		}

		children, _ := parent["children"].([]map[string]any)
		parent["children"] = append(children, payload)
	}

	return roots
}
