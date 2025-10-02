// Package api содержит HTTP-обработчики и вспомогательные функции для REST API планировщика.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"todo/pkg/db"
)

// taskHandler — общий роутер для пути /api/task.
// Внутри по HTTP-методу вызываются соответствующие под-обработчики:
// POST (add), GET (get by id), PUT (update), DELETE (remove).
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		writeJSON(w, map[string]string{"error": "method not allowed"})
	}
}

// addTaskHandler обрабатывает POST /api/task.
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	t := new(db.Task)
	if err := json.NewDecoder(r.Body).Decode(t); err != nil {
		writeError(w, http.StatusBadRequest, "json parse error")
		return
	}
	if t.Title == "" {
		writeError(w, http.StatusBadRequest, "empty title")
		return
	}
	if err := checkDate(t); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, err := db.AddTask(t)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db insert error")
		return
	}
	writeJSON(w, map[string]string{"id": fmt.Sprint(id)})
}

// getTaskHandler — GET /api/task?id=<число>
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "no id")
		return
	}
	t, err := db.GetTask(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeJSON(w, t) // db.Task сериализуется напрямую (id -> string через тег)
}

// updateTaskHandler — PUT /api/task
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	in := new(db.Task)
	if err := json.NewDecoder(r.Body).Decode(in); err != nil {
		writeError(w, http.StatusBadRequest, "json parse error")
		return
	}
	if in.ID <= 0 {
		writeError(w, http.StatusBadRequest, "bad id")
		return
	}
	if in.Title == "" {
		writeError(w, http.StatusBadRequest, "empty title")
		return
	}
	if err := checkDate(in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := db.UpdateTask(in); err != nil {
		writeError(w, http.StatusNotFound, "update error")
		return
	}
	writeJSON(w, map[string]any{})
}

// checkDate — единая проверка/нормализация даты и repeat.
// Приводит пустую дату к сегодняшней, проверяет формат, сдвигает дату в будущее.
func checkDate(tk *db.Task) error {
	now := time.Now()
	y, m, d := now.Date()
	now = time.Date(y, m, d, 0, 0, 0, 0, time.Local)

	if tk.Date == "" {
		tk.Date = now.Format(dateFmt)
	}

	td, err := time.Parse(dateFmt, tk.Date)
	if err != nil {
		return fmt.Errorf("bad date format")
	}

	var next string
	if tk.Repeat != "" {
		next, err = NextDate(now, tk.Date, tk.Repeat)
		if err != nil {
			return fmt.Errorf("bad repeat")
		}
	}

	// если дата в прошлом — берём сегодня (без repeat) или следующую (с repeat)
	if td.Before(now) {
		if tk.Repeat == "" {
			tk.Date = now.Format(dateFmt)
		} else {
			tk.Date = next
		}
	}
	return nil
}

// deleteTaskHandler — DELETE /api/task?id=...
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "no id")
		return
	}
	if err := db.DeleteTask(id); err != nil {
		writeError(w, http.StatusNotFound, "delete error")
		return
	}
	writeJSON(w, map[string]any{})
}

// taskDoneHandler — POST /api/task/done?id=...
func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "no id")
		return
	}
	t, err := db.GetTask(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if strings.TrimSpace(t.Repeat) == "" {
		if err := db.DeleteTask(id); err != nil {
			writeError(w, http.StatusNotFound, "delete error")
			return
		}
		writeJSON(w, map[string]any{})
		return
	}
	now := time.Now()
	y, m, d := now.Date()
	now = time.Date(y, m, d, 0, 0, 0, 0, time.Local)

	next, err := NextDate(now, t.Date, t.Repeat)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad repeat")
		return
	}
	if err := db.UpdateDate(next, id); err != nil {
		writeError(w, http.StatusNotFound, "update error")
		return
	}
	writeJSON(w, map[string]any{})
}
