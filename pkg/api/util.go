// Package api: утилиты для API-ответов.
package api

import (
	"encoding/json"
	"net/http"
)

// writeJSON сериализует данные в JSON и пишет в ответ с нужным Content-Type.
// Используется всеми обработчиками API.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	_ = json.NewEncoder(w).Encode(v)
}
