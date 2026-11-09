package health

// Health represents the health status of the application
type Health struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
