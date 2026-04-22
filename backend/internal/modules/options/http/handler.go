package http

import (
	"net/http"

	"credit-management/backend/internal/httpapi/response"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) GetOrganizationTree(c *gin.Context) {
	response.Success(c, http.StatusOK, gin.H{
		"items": []gin.H{
			{
				"id":        "school-01",
				"name":      "Demo University",
				"node_type": "school",
				"children": []gin.H{
					{
						"id":        "college-01",
						"name":      "Engineering College",
						"node_type": "college",
						"children": []gin.H{
							{
								"id":        "major-01",
								"name":      "Software Engineering",
								"node_type": "major",
								"children": []gin.H{
									{
										"id":        "class-01",
										"name":      "2024 Class 1",
										"node_type": "class",
									},
								},
							},
						},
					},
				},
			},
		},
	})
}

func (h *Handler) GetGrades(c *gin.Context) {
	response.OK(c, gin.H{
		"items": []gin.H{
			{"value": "2024", "label": "2024"},
			{"value": "2025", "label": "2025"},
			{"value": "2026", "label": "2026"},
		},
	})
}

func (h *Handler) GetAcademicYears(c *gin.Context) {
	response.OK(c, gin.H{
		"items": []gin.H{
			{"value": "2024-2025", "label": "2024-2025"},
			{"value": "2025-2026", "label": "2025-2026"},
			{"value": "2026-2027", "label": "2026-2027"},
		},
	})
}

func (h *Handler) GetSemesters(c *gin.Context) {
	response.OK(c, gin.H{
		"items": []gin.H{
			{"value": "spring", "label": "Spring"},
			{"value": "autumn", "label": "Autumn"},
		},
	})
}

func (h *Handler) GetTeacherTitles(c *gin.Context) {
	response.OK(c, gin.H{
		"items": []gin.H{
			{"value": "assistant", "label": "Assistant"},
			{"value": "lecturer", "label": "Lecturer"},
			{"value": "associate_professor", "label": "Associate Professor"},
			{"value": "professor", "label": "Professor"},
		},
	})
}
