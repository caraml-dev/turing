package service

import "github.com/jinzhu/gorm"

type PaginationQuery interface {
	Page() int
	PageSize() int
}

type paginationQuery struct {
	page     int `schema:"page"`
	pageSize int `schema:"page_size"`
}

func (q paginationQuery) Page() int {
	return q.page
}

func (q paginationQuery) PageSize() int {
	return q.pageSize
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
		offset := (q.Page() - 1) * q.PageSize()
		return db.Offset(offset).Limit(q.PageSize())
	}
}
