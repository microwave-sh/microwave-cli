package client

import "context"

type SortDirective struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type SearchRequest struct {
	Filter map[string]map[string]any `json:"filter,omitempty"`
	Sort   []SortDirective           `json:"sort,omitempty"`
	Limit  int                       `json:"limit,omitempty"`
	Cursor string                    `json:"cursor,omitempty"`
}

type SearchResponse[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
}

func Search[T any](ctx context.Context, c *Client, path string, req SearchRequest) (*SearchResponse[T], error) {
	var page SearchResponse[T]
	if err := c.Do(ctx, "POST", path+"/search", req, &page); err != nil {
		return nil, err
	}
	return &page, nil
}
