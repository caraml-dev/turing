package service

import "github.com/jinzhu/gorm"

type PaginationOptions struct {
	Page     *int `schema:"page" validate:"min=1"`
	PageSize *int `schema:"page_size" validate:"min=1"`
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
