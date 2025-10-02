// Package api: регистрация HTTP-маршрутов API.
// Здесь связываем URL с обработчиками и навешиваем middleware (auth).
package api

import "net/http"

// Init регистрирует все маршруты API в стандартном mux.
// /api/signin — вход (выдача JWT), остальные — защищённые (auth(...)).
func Init() {
	setPasswordFromEnv() // ← добавили

	http.HandleFunc("/api/signin", signinHandler)
	http.HandleFunc("/api/task", auth(taskHandler))
	http.HandleFunc("/api/tasks", auth(tasksHandler))
	http.HandleFunc("/api/task/done", auth(taskDoneHandler))
	http.HandleFunc("/api/nextdate", nextDateHandler)
}
