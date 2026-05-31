package response

import (
	"encoding/json"
	"net/http"
)

// Envelope is the consistent JSON shape returned by every endpoint.
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func write(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}

// OK writes a 200 with data.
func OK(w http.ResponseWriter, data interface{}) {
	write(w, http.StatusOK, Envelope{Success: true, Data: data})
}

// Created writes a 201 with data.
func Created(w http.ResponseWriter, data interface{}) {
	write(w, http.StatusCreated, Envelope{Success: true, Data: data})
}

// Paginated writes a 200 with data plus pagination meta.
func Paginated(w http.ResponseWriter, data interface{}, meta interface{}) {
	write(w, http.StatusOK, Envelope{Success: true, Data: data, Meta: meta})
}

// NoContent writes a 204.
func NoContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }

// Error writes an error envelope with the given status and code.
func Error(w http.ResponseWriter, status int, code, message string) {
	write(w, status, Envelope{Success: false, Error: &ErrorBody{Code: code, Message: message}})
}

// ValidationError writes a 422 with per-field messages.
func ValidationError(w http.ResponseWriter, fields map[string]string) {
	write(w, http.StatusUnprocessableEntity, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: "validation_error", Message: "Dados inválidos", Fields: fields},
	})
}
