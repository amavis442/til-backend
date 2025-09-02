package handler

type Response[T any] struct {
	Items  T      `json:"items,omitempty"`
	Total  int64  `json:"total,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Error  string `json:"error,omitempty"`
}

func ErrorResponse(msg string) Response[any] {
	return Response[any]{Error: msg}
}
