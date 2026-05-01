package service

import (
	"context"
	"errors"
	"testing"
)

func TestServiceReturnsOrganizationTreeAndOptionListsFromRepository(t *testing.T) {
	repo := fakeRepository{
		nodes: []OrganizationNode{
			{ID: "school-01", Name: "Demo University", NodeType: "school", Path: "school-01", Level: 1, SortOrder: 1},
			{ID: "college-01", ParentID: "school-01", Name: "Engineering College", NodeType: "college", Path: "school-01/college-01", Level: 2, SortOrder: 1},
			{ID: "major-01", ParentID: "college-01", Name: "Software Engineering", NodeType: "major", Path: "school-01/college-01/major-01", Level: 3, SortOrder: 1},
			{ID: "class-01", ParentID: "major-01", Name: "2024 Class 1", NodeType: "class", Path: "school-01/college-01/major-01/class-01", Level: 4, SortOrder: 1},
		},
		itemsByGroup: map[string][]OptionItem{
			"academic_years": {
				{Value: "2025-2026", Label: "2025-2026"},
			},
		},
	}
	service := NewService(repo)

	tree, err := service.OrganizationTree(context.Background())
	if err != nil {
		t.Fatalf("organization tree: %v", err)
	}
	if len(tree) != 1 {
		t.Fatalf("expected one root node, got %d", len(tree))
	}
	if tree[0]["id"] != "school-01" || tree[0]["node_type"] != "school" {
		t.Fatalf("unexpected root node: %#v", tree[0])
	}

	years, err := service.OptionList(context.Background(), "academic_years")
	if err != nil {
		t.Fatalf("option list: %v", err)
	}
	if len(years) != 1 {
		t.Fatalf("expected one academic year, got %d", len(years))
	}
	if years[0]["value"] != "2025-2026" || years[0]["label"] != "2025-2026" {
		t.Fatalf("unexpected academic year item: %#v", years[0])
	}
}

func TestServicePropagatesRepositoryErrors(t *testing.T) {
	service := NewService(fakeRepository{treeErr: errors.New("boom")})

	if _, err := service.OrganizationTree(context.Background()); err == nil {
		t.Fatal("expected organization tree error")
	}
	if _, err := service.OptionList(context.Background(), "grades"); err == nil {
		t.Fatal("expected option list error")
	}
}

type fakeRepository struct {
	nodes        []OrganizationNode
	itemsByGroup map[string][]OptionItem
	treeErr      error
	itemsErr     error
}

func (f fakeRepository) ListOrganization(ctx context.Context) ([]OrganizationNode, error) {
	if f.treeErr != nil {
		return nil, f.treeErr
	}
	return f.nodes, nil
}

func (f fakeRepository) ListOptionItems(ctx context.Context, groupKey string) ([]OptionItem, error) {
	if f.itemsErr != nil || f.treeErr != nil {
		if f.itemsErr != nil {
			return nil, f.itemsErr
		}
		return nil, f.treeErr
	}
	return f.itemsByGroup[groupKey], nil
}
