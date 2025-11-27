package dto

import "github.com/gin-gonic/gin"

type PaginatedResponse[T any] struct {
	Items       []T   `json:"items"`
	Total       int64 `json:"total"`
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

func BuildPaginatedResponse[T any](items []T, total int64, page, limit int) gin.H {
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages
	hasPrevious := page > 1

	return gin.H{
		"items":        items,
		"total":        total,
		"page":         page,
		"limit":        limit,
		"total_pages":  totalPages,
		"has_next":     hasNext,
		"has_previous": hasPrevious,
	}
}
