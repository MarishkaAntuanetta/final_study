// Package api: обработчик списка задач с опциональным поиском.
// GET /api/tasks[?search=...][&limit=N]
package api

import (
	"net/http"
	"strconv"

	"todo/pkg/db"
)

// tasksResp — форма ответа: {"tasks":[...]}.
// Элементы — напрямую db.Task (id сериализуется строкой через тег json:",string").
type tasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// tasksHandler — обрабатывает GET /api/tasks.
// Поддерживает ограничение limit и поиск search
// (подстрока в title/comment или дата 02.01.2006).
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, map[string]string{"error": "method not allowed"})
		return
	}

	search := r.URL.Query().Get("search")

	// дефолтный лимит берём из константы пакета
	limit := defaultTasksLimit
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
		}
	}

	items, err := db.Tasks(limit, search)
	if err != nil {
		writeJSON(w, map[string]string{"error": "db select error"})
		return
	}

	writeJSON(w, tasksResp{Tasks: items})
}
