package handler

type ErrorResponse struct {
	Message string `json:"message"`
}

type Response struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}
