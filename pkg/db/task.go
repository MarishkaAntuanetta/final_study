// Package db: модели и операции с задачами (CRUD) поверх SQLite.
package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Task описывает одну задачу из таблицы scheduler.
// В БД все поля (кроме id) текстовые; в коде id удобнее хранить как int64.
type Task struct {
	ID      int64  `json:"id,string" db:"id"`
	Date    string `json:"date" db:"date"`
	Title   string `json:"title" db:"title"`
	Comment string `json:"comment" db:"comment"`
	Repeat  string `json:"repeat" db:"repeat"`
}

// AddTask вставляет новую задачу в таблицу scheduler и возвращает её идентификатор.
func AddTask(task *Task) (int64, error) {
	const q = `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB.Exec(q, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Tasks возвращает список задач, отсортированных по дате (возрастание).
// Поддерживает простой поиск:
//   - search == ""        → просто LIMIT
//   - search как 02.01.2006 → фильтр по точной дате (конвертируем в 20060102)
//   - иначе               → LIKE по title и comment
func Tasks(limit int, search string) ([]*Task, error) {
	if limit <= 0 {
		limit = 50
	}

	var rows *sql.Rows
	var err error

	if search == "" {
		rows, err = DB.Query(
			`SELECT id, date, title, comment, repeat
			 FROM scheduler
			 ORDER BY date
			 LIMIT ?`, limit)
	} else {
		// Пытаемся распознать строку как дату 02.01.2006.
		if t, e := time.Parse("02.01.2006", search); e == nil {
			dateStr := t.Format("20060102")
			rows, err = DB.Query(
				`SELECT id, date, title, comment, repeat
				 FROM scheduler
				 WHERE date = ?
				 ORDER BY date
				 LIMIT ?`, dateStr, limit)
		} else {
			// Иначе ищем подстроку в title/comment через LIKE (регистр-чувствительный).
			p := "%" + search + "%"
			rows, err = DB.Query(
				`SELECT id, date, title, comment, repeat
				 FROM scheduler
				 WHERE title LIKE ? OR comment LIKE ?
				 ORDER BY date
				 LIMIT ?`, p, p, limit)
		}
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Task
	for rows.Next() {
		t := &Task{}
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	// Чтобы JSON-маршалинг выдавал "tasks": [] (а не null) при отсутствии данных.
	if out == nil {
		out = make([]*Task, 0)
	}
	return out, nil
}

// GetTask возвращает одну задачу по её строковому идентификатору (например, "185").
// Если записи нет — возвращает ошибку вида "task not found".
func GetTask(id string) (*Task, error) {
	row := DB.QueryRow(
		`SELECT id, date, title, comment, repeat
		 FROM scheduler
		 WHERE id = ?`, id)

	t := &Task{}
	if err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}
	return t, nil
}

// UpdateTask обновляет все основные поля задачи по её ID.
func UpdateTask(task *Task) error {
	res, err := DB.Exec(
		`UPDATE scheduler
		 SET date = ?, title = ?, comment = ?, repeat = ?
		 WHERE id = ?`,
		task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("incorrect id for updating task")
	}
	return nil
}

// DeleteTask удаляет задачу по её идентификатору.
// Если ни одна строка не затронута — возвращает ошибку "task not found".
func DeleteTask(id string) error {
	res, err := DB.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

// UpdateDate обновляет только поле date у задачи с заданным id.
// Полезно при отметке задачи "выполненной" с пересчётом следующей даты.
func UpdateDate(next string, id string) error {
	res, err := DB.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, next, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}
