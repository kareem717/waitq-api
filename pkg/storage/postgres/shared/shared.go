package shared

import (
	"github.com/uptrace/bun"
)

func ExcludeInsertColumns(query *bun.InsertQuery) *bun.InsertQuery {
	query.ExcludeColumn("created_at", "updated_at", "deleted_at", "id")
	return query
}

func ExcludeUpdateColumns(query *bun.UpdateQuery) *bun.UpdateQuery {
	query.ExcludeColumn("created_at", "updated_at", "deleted_at", "id")
	return query
}

type GetManyRequest struct {
	IncludeDeleted bool `json:"includeDeleted" default:"false" required:"false"`
}

type PaginationRequest struct {
	Page     int `json:"page" default:"1" min:"1" required:"false"`
	PageSize int `json:"pageSize" default:"10" min:"1" max:"100" required:"false"`
	GetManyRequest
}

type PaginationResponse struct {
	Data     []interface{}
	Total    int
	Page     int
	PageSize int
}
