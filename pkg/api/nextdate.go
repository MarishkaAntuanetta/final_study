// Package api: расчёт следующей даты для повторяющихся задач.
// Здесь хранятся NextDate(...) и nextDateHandler для /api/nextdate.
package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// afterNow возвращает true, если дата d строго больше now (сравнение по дню).
func afterNow(d, now time.Time) bool {
	yd, md, dd := d.Date()
	yn, mn, dn := now.Date()
	d = time.Date(yd, md, dd, 0, 0, 0, 0, time.Local)
	now = time.Date(yn, mn, dn, 0, 0, 0, 0, time.Local)
	return d.After(now)
}

// NextDate — базовая логика повторений.
// Поддержаны правила: "y" (ежегодно), "d N" (через N дней, 1..400).
// Остальные форматы пока считаются неподдерживаемыми (ошибка).
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if strings.TrimSpace(repeat) == "" {
		return "", errors.New("repeat is empty")
	}
	start, err := time.Parse(dateFmt, dstart)
	if err != nil {
		return "", err
	}

	parts := strings.Fields(repeat)
	switch parts[0] {
	case "y":
		d := start
		for {
			d = d.AddDate(1, 0, 0)
			if afterNow(d, now) {
				return d.Format(dateFmt), nil
			}
		}
	case "d":
		if len(parts) != 2 {
			return "", errors.New("bad d format")
		}
		n, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", errors.New("bad d number")
		}
		if n <= 0 || n > 400 {
			return "", errors.New("bad d interval")
		}
		d := start
		for {
			d = d.AddDate(0, 0, n)
			if afterNow(d, now) {
				return d.Format(dateFmt), nil
			}
		}
	default:
		return "", errors.New("unsupported repeat")
	}
}

// nextDateHandler — GET /api/nextdate?now=20060102&date=20060102&repeat=...
// Возвращает дату следующего выполнения (строкой) или текст "error".
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	nowStr := strings.TrimSpace(r.FormValue("now"))
	dateStr := strings.TrimSpace(r.FormValue("date"))
	repeat := strings.TrimSpace(r.FormValue("repeat"))

	var now time.Time
	var err error
	if nowStr == "" {
		t := time.Now()
		y, m, d := t.Date()
		now = time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	} else {
		now, err = time.Parse(dateFmt, nowStr)
		if err != nil {
			fmt.Fprintln(w, "bad now")
			return
		}
	}

	next, err := NextDate(now, dateStr, repeat)
	if err != nil {
		// корректная обработка: 400 + JSON с ошибкой
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}
	fmt.Fprintln(w, next)
}
