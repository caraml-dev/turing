package service

import (
	"math"

	"gorm.io/gorm"
)

type PaginationOptions struct {
	Page     *int `schema:"page" validate:"omitempty,min=1"`
	PageSize *int `schema:"page_size" validate:"omitempty,min=1"`
}

type Paging struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Pages int `json:"pages"`
}

type PaginatedResults struct {
	Results interface{} `json:"results"`
	Paging  Paging      `json:"paging"`
}

func PaginationScope(options PaginationOptions) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if options.PageSize != nil {
			db = db.Limit(*options.PageSize)
			if options.Page != nil {
				db = db.Offset((*options.Page - 1) * *options.PageSize)
			}
		}
		return db
	}
}

// createPaginatedResults is a helper function that helps to create paginated results
func createPaginatedResults(options PaginationOptions, count int, results interface{}) *PaginatedResults {
	page := 1
	totalPages := 1
	if options.Page != nil && options.PageSize != nil {
		page = int(math.Max(1, float64(*options.Page)))
		totalPages = int(math.Ceil(float64(count) / float64(*options.PageSize)))
	}

	return &PaginatedResults{
		Results: results,
		Paging: Paging{
			Total: count,
			Page:  page,
			Pages: totalPages,
		},
	}
}
