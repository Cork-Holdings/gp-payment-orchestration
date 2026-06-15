package common

type PageInfo struct {
	TotalPages uint `json:"total_pages,omitempty"`
	Count      uint `json:"count,omitempty"`
	Page       uint `json:"page,omitempty"`
	Size       uint `json:"size,omitempty"`
} //	@name	PaginationInfo
type Response struct {
	Message string    `json:"message"`
	Status  int       `json:"status"`
	Data    any       `json:"data,omitempty"`
	Code    int       `json:"code,omitempty"`
	Meta    *PageInfo `json:"meta,omitempty" swaggerignore:"true"`
} //	@name	StandardResponse
