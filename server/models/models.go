package models

type AppError struct {
	Message    string
	StatusCode int
}

func (e AppError) Error() string {
	return e.Message
}

func NewAppError(message string, statusCode int) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: statusCode,
	}
}

type DataItem struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// DataResponse represents the response for the data endpoint
type DataResponse struct {
	Data  []DataItem `json:"data"`
	Count int        `json:"count"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UsersResponse struct {
	Users []User `json:"users"`
	Count int    `json:"count"`
}

type ProcessRequest struct {
	Items []string          `json:"items"`
	Args  map[string]string `json:"args"`
}

type ProcessResponse struct {
	Success   bool   `json:"success"`
	ProcessID int    `json:"process_id"`
	Message   string `json:"message"`
}
