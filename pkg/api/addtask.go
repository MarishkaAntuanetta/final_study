// Package api содержит HTTP-обработчики и вспомогательные функции для REST API планировщика.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
// Принимает JSON c полями date, title, comment, repeat.
// Валидирует данные, нормализует дату (через checkDate) и добавляет запись в БД.
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, map[string]string{"error": "read body error"})
		return
	}
	defer r.Body.Close()

	var t db.Task
	if err := json.Unmarshal(body, &t); err != nil {
		writeJSON(w, map[string]string{"error": "json parse error"})
		return
	}
	if t.Title == "" {
		writeJSON(w, map[string]string{"error": "empty title"})
		return
	}
	if err := checkDate(&t); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}
	id, err := db.AddTask(&t)
	if err != nil {
		writeJSON(w, map[string]string{"error": "db insert error"})
		return
	}
	writeJSON(w, map[string]string{"id": fmt.Sprint(id)})
}

// getTaskHandler обрабатывает GET /api/task?id=<число>.
// Возвращает полную информацию о задаче в виде JSON.
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "no id"})
		return
	}
	t, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, map[string]string{"error": "task not found"})
		return
	}
	resp := map[string]string{
		"id":      strconv.FormatInt(t.ID, 10),
		"date":    t.Date,
		"title":   t.Title,
		"comment": t.Comment,
		"repeat":  t.Repeat,
	}
	writeJSON(w, resp)
}

// updateTaskHandler обрабатывает PUT /api/task.
// Принимает JSON (как при добавлении) + поле id. Валидирует и обновляет запись.
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, map[string]string{"error": "read body error"})
		return
	}
	defer r.Body.Close()

	type taskIn struct {
		ID      string `json:"id"`
		Date    string `json:"date"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
		Repeat  string `json:"repeat"`
	}
	var in taskIn
	if err := json.Unmarshal(body, &in); err != nil {
		writeJSON(w, map[string]string{"error": "json parse error"})
		return
	}
	if in.ID == "" {
		writeJSON(w, map[string]string{"error": "no id"})
		return
	}
	if in.Title == "" {
		writeJSON(w, map[string]string{"error": "empty title"})
		return
	}
	idNum, err := strconv.ParseInt(in.ID, 10, 64)
	if err != nil || idNum <= 0 {
		writeJSON(w, map[string]string{"error": "bad id"})
		return
	}

	t := db.Task{
		ID:      idNum,
		Date:    in.Date,
		Title:   in.Title,
		Comment: in.Comment,
		Repeat:  in.Repeat,
	}
	if err := checkDate(&t); err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}
	if err := db.UpdateTask(&t); err != nil {
		writeJSON(w, map[string]string{"error": "update error"})
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

// deleteTaskHandler обрабатывает DELETE /api/task?id=<число>.
// Удаляет задачу по идентификатору.
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "no id"})
		return
	}
	if err := db.DeleteTask(id); err != nil {
		writeJSON(w, map[string]string{"error": "delete error"})
		return
	}
	writeJSON(w, map[string]any{})
}

// taskDoneHandler обрабатывает POST /api/task/done?id=<число>.
// Если repeat пустой — удаляет задачу. Если периодическая — вычисляет следующую дату и
// обновляет только поле date.
func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, map[string]string{"error": "method not allowed"})
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "no id"})
		return
	}

	// читаем текущую задачу
	t, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, map[string]string{"error": "task not found"})
		return
	}

	// одноразовая — просто удалить
	if strings.TrimSpace(t.Repeat) == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJSON(w, map[string]string{"error": "delete error"})
			return
		}
		writeJSON(w, map[string]any{})
		return
	}

	// периодическая — посчитать следующую дату и обновить
	now := time.Now()
	y, m, d := now.Date()
	now = time.Date(y, m, d, 0, 0, 0, 0, time.Local)

	next, err := NextDate(now, t.Date, t.Repeat)
	if err != nil {
		writeJSON(w, map[string]string{"error": "bad repeat"})
		return
	}
	if err := db.UpdateDate(next, id); err != nil {
		writeJSON(w, map[string]string{"error": "update error"})
		return
	}
	writeJSON(w, map[string]any{})
}
