package response

import (
	"encoding/json"
	"log"
	"net/http"
)

// Body é o envelope padrão de todas as respostas da API.
//
//	sucesso: { "status": true,  "statusCode": 200, "data": {...} }
//	erro:    { "status": false, "statusCode": 422, "error": "mensagem", "details": {...} }
type Body struct {
	Status     bool   `json:"status"`
	StatusCode int    `json:"statusCode"`
	Data       any    `json:"data,omitempty"`
	Error      string `json:"error,omitempty"`
	Details    any    `json:"details,omitempty"`
}

// Paginated é o formato padrão para listagens paginadas (vai dentro de "data").
type Paginated struct {
	Items   any   `json:"items"`
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"perPage"`
}

func Success(w http.ResponseWriter, statusCode int, data any) {
	write(w, statusCode, Body{
		Status:     true,
		StatusCode: statusCode,
		Data:       data,
	})
}

func Error(w http.ResponseWriter, statusCode int, message string) {
	ErrorWithDetails(w, statusCode, message, nil)
}

func ErrorWithDetails(w http.ResponseWriter, statusCode int, message string, details any) {
	write(w, statusCode, Body{
		Status:     false,
		StatusCode: statusCode,
		Error:      message,
		Details:    details,
	})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func write(w http.ResponseWriter, statusCode int, body Body) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("response: erro ao serializar JSON: %v", err)
	}
}
