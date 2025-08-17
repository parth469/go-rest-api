package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SuccessResponse[T any] struct {
	Timestamp   string        `json:"timestamp"`
	Data        T             `json:"data"`
	Message     string        `json:"message"`
	Status      int           `json:"status"`
	ProcessTime time.Duration `json:"processTime,omitempty"`
}

type ErrorResponse struct {
	Timestamp string `json:"timestamp"`
	Method    string `json:"method"`
	Status    int    `json:"status"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"` // hide if empty
}

func SuccessWriter[T any](w http.ResponseWriter, rowData T, r *http.Request) {

	var status int
	message := "Request processed successfully"

	switch r.Method {
	case http.MethodGet:
		message = "Resource fetched successfully"
		status = http.StatusOK
	case http.MethodPost:
		message = "Resource created successfully"
		status = http.StatusCreated
	case http.MethodPut, http.MethodPatch:
		message = "Resource updated successfully"
		status = http.StatusOK
	case http.MethodDelete:
		message = "Resource deleted successfully"
		status = http.StatusOK
	default:
		status = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := SuccessResponse[T]{
		Data:      rowData,
		Timestamp: time.Now().Format(time.RFC3339), // ISO-8601
		Message:   message,
		Status:    status,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		ErrorWriter(w, r, http.StatusBadRequest, fmt.Errorf("failed to encode response"))
		return
	}

	fmt.Fprintln(w, response)
}
func ErrorWriter(w http.ResponseWriter, r *http.Request, status int, err error) {
	statusMessageMap := map[int]string{
		// 4xx Client Errors
		http.StatusBadRequest:          "Invalid request parameters or malformed request",
		http.StatusUnauthorized:        "Authentication required or invalid credentials",
		http.StatusForbidden:           "You don't have permission to access this resource",
		http.StatusNotFound:            "The requested resource was not found",
		http.StatusMethodNotAllowed:    "HTTP method not allowed for this endpoint",
		http.StatusConflict:            "Resource conflict or duplicate entry",
		http.StatusUnprocessableEntity: "Request validation failed",
		http.StatusTooManyRequests:     "Too many requests, please try again later",

		// 5xx Server Errors
		http.StatusInternalServerError: "Internal server error",
		http.StatusNotImplemented:      "Feature not implemented",
		http.StatusServiceUnavailable:  "Service temporarily unavailable",
	}

	message, ok := statusMessageMap[status]
	if !ok {
		message = "An unexpected error occurred. Please try again later or contact support."
	}

	// Don’t leak internal error details for 5xx
	errorMessage := ""
	if status < 500 && err != nil {
		errorMessage = err.Error()
	}

	response := ErrorResponse{
		Timestamp: time.Now().Format(time.RFC3339), // ISO-8601
		Method:    r.Method,
		Status:    status,
		Message:   message,
		Error:     errorMessage,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		http.Error(w, "failed to encode error response", http.StatusInternalServerError)
	}
}
