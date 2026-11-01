package response

// Response represents a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success creates a successful response payload
func Success(message string, data ...interface{}) Response {
	resp := Response{
		Success: true,
		Message: message,
	}
	if len(data) > 0 {
		resp.Data = data[0]
	}
	return resp
}

// Error creates an error response payload
func Error(message string) Response {
	return Response{
		Success: false,
		Message: message,
	}
}
