package v1

type paginatedResponse struct {
	Data interface{}        `json:"data"`
	Meta paginationMetaData `json:"meta"`
}

type paginationMetaData struct {
	Total     int `json:"total"`
	Displayed int `json:"displayed"`
}

func newPaginatedResponse(data interface{}, total, displayed int) paginatedResponse {
	return paginatedResponse{
		Data: data,
		Meta: paginationMetaData{
			Total:     total,
			Displayed: displayed,
		},
	}
}
