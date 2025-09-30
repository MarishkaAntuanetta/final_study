// Package api: обработчик списка задач с опциональным поиском.
// GET /api/tasks[?search=...][&limit=N]
package api

import (
	"net/http"
	"strconv"

	"todo/pkg/db"
)

// taskOut — форма одной задачи в ответе списка (все поля строками).
type taskOut struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// tasksResp — форма всего ответа: {"tasks":[...]}.
type tasksResp struct {
	Tasks []taskOut `json:"tasks"`
}

// tasksHandler — обрабатывает GET /api/tasks.
// Поддерживает ограничение limit и поиск search (подстрока в title/comment или дата 02.01.2006).
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, map[string]string{"error": "method not allowed"})
		return
	}

	search := r.URL.Query().Get("search")
	limit := 50
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

	out := make([]taskOut, 0, len(items))
	for _, t := range items {
		out = append(out, taskOut{
			ID:      strconv.FormatInt(t.ID, 10),
			Date:    t.Date,
			Title:   t.Title,
			Comment: t.Comment,
			Repeat:  t.Repeat,
		})
	}

	writeJSON(w, tasksResp{Tasks: out})
}
