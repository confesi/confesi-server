package admin

import (
	"confesi/db"
	"confesi/lib/response"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *handler) handleListFeedback(c *gin.Context) {

	// get query params
	pageStr := c.Query("page")
	pageSizeStr := c.Query("limit")

	// Error Check
	page, err := strconv.Atoi(pageStr)

	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid page").Send(c)
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		response.New(http.StatusBadRequest).Err("invalid limit").Send(c)
		return
	}

	switch {
	case pageSize <= 0:
		pageSize = 10
	case page <= 0:
		page = 1
	case pageSize > 999:
		page = 10
	}

	pagination := Pagination{
		Limit: pageSize,
		Page:  page,
		Sort:  "id desc",
	}

	var feedback []db.Feedback

	h.db.Scopes(paginate(feedback, &pagination, h.db)).Find(&feedback)
	pagination.Rows = feedback
	if len(feedback) == 0 {
		response.New(http.StatusNotFound).Err("no feedback found").Send(c)
		return
	}

	// if all goes well, send 200
	response.New(http.StatusOK).Val(pagination).Send(c)
	return
}

//https://dev.to/rafaelgfirmino/pagination-using-gorm-scopes-3k5f

type Pagination struct {
	Limit      int         `json:"limit,omitempty;query:limit"`
	Page       int         `json:"page,omitempty;query:page"`
	Sort       string      `json:"sort,omitempty;query:sort"`
	TotalRows  int64       `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
	Rows       interface{} `json:"rows"`
}

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *Pagination) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = 10
	}
	return p.Limit
}

func (p *Pagination) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}

func (p *Pagination) GetSort() string {
	if p.Sort == "" {
		p.Sort = "id desc"
	}
	return p.Sort
}

func paginate(value interface{}, pagination *Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	db.Model(value).Count(&totalRows)

	pagination.TotalRows = totalRows
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}
