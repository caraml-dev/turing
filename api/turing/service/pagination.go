package service

import "github.com/jinzhu/gorm"

type PaginationQuery struct {
	Page     int `schema:"page"`
	PageSize int `schema:"page_size"`
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

func Paginate(q PaginationQuery) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (q.Page - 1) * q.PageSize
		return db.Offset(offset).Limit(q.PageSize)
	}
}
