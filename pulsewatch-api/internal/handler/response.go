package handler

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// JSON writes any data as indented JSON with the specified status code.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// JSONError writes a standardized error payload.
func JSONError(w http.ResponseWriter, status int, errCode, message string) {
	JSON(w, status, ErrorResponse{
		Error:   errCode,
		Message: message,
	})
}

// JSONMessage writes a simple confirmation message payload.
func JSONMessage(w http.ResponseWriter, status int, message string) {
	JSON(w, status, MessageResponse{
		Message: message,
	})
}
