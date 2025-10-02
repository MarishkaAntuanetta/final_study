// Package api: утилиты для API-ответов.
package api

import (
	"encoding/json"
	"net/http"
)

// writeJSON сериализует данные в JSON и пишет 200 OK.
func writeJSON(w http.ResponseWriter, v any) {
	writeJSONStatus(w, http.StatusOK, v)
}

// writeJSONStatus — JSON + явный статус.
func writeJSONStatus(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError — короткий хелпер для ошибок.
func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSONStatus(w, code, map[string]string{"error": msg})
}
