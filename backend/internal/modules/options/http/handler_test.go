package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	optionsservice "credit-management/backend/internal/modules/options/service"

	"github.com/gin-gonic/gin"
)

func TestGetOrganizationTreePreservesFrontendShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(optionsservice.NewService(handlerRepo{
		nodes: []optionsservice.OrganizationNode{
			{ID: "school-01", Name: "Demo University", NodeType: "school", Path: "school-01", Level: 1, SortOrder: 1},
			{ID: "college-01", ParentID: "school-01", Name: "Engineering College", NodeType: "college", Path: "school-01/college-01", Level: 2, SortOrder: 1},
			{ID: "major-01", ParentID: "college-01", Name: "Software Engineering", NodeType: "major", Path: "school-01/college-01/major-01", Level: 3, SortOrder: 1},
			{ID: "class-01", ParentID: "major-01", Name: "2024 Class 1", NodeType: "class", Path: "school-01/college-01/major-01/class-01", Level: 4, SortOrder: 1},
		},
	}))

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/options/organization/tree", nil)

	handler.GetOrganizationTree(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Code string `json:"code"`
		Data struct {
			Items []map[string]any `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if payload.Code != "OK" {
		t.Fatalf("unexpected code: %s", payload.Code)
	}
	if len(payload.Data.Items) != 1 {
		t.Fatalf("expected one root item, got %d", len(payload.Data.Items))
	}
	root := payload.Data.Items[0]
	if root["id"] != "school-01" || root["name"] != "Demo University" || root["node_type"] != "school" {
		t.Fatalf("unexpected root node shape: %#v", root)
	}
	if _, ok := root["children"].([]any); !ok {
		t.Fatalf("expected children array, got %#v", root["children"])
	}
}

func TestGetGradesReturnsOptionItemsShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(optionsservice.NewService(handlerRepo{
		itemsByGroup: map[string][]optionsservice.OptionItem{
			"grades": {
				{Value: "2024", Label: "2024"},
				{Value: "2025", Label: "2025"},
			},
		},
	}))

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/options/grades", nil)

	handler.GetGrades(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var payload struct {
		Data struct {
			Items []map[string]string `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v\nbody=%s", err, recorder.Body.String())
	}
	if len(payload.Data.Items) != 2 {
		t.Fatalf("expected two grade items, got %d", len(payload.Data.Items))
	}
	if payload.Data.Items[0]["value"] != "2024" || payload.Data.Items[0]["label"] != "2024" {
		t.Fatalf("unexpected grade item: %#v", payload.Data.Items[0])
	}
}

func TestGetTeacherTitlesReturnsInternalErrorOnServiceFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(optionsservice.NewService(handlerRepo{
		itemsErr: errors.New("db down"),
	}))

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/options/teacher-titles", nil)

	handler.GetTeacherTitles(ctx)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

type handlerRepo struct {
	nodes        []optionsservice.OrganizationNode
	itemsByGroup map[string][]optionsservice.OptionItem
	treeErr      error
	itemsErr     error
}

func (h handlerRepo) ListOrganization(ctx context.Context) ([]optionsservice.OrganizationNode, error) {
	if h.treeErr != nil {
		return nil, h.treeErr
	}
	return h.nodes, nil
}

func (h handlerRepo) ListOptionItems(ctx context.Context, groupKey string) ([]optionsservice.OptionItem, error) {
	if h.itemsErr != nil {
		return nil, h.itemsErr
	}
	return h.itemsByGroup[groupKey], nil
}
