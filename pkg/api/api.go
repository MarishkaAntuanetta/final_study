// Package api: регистрация HTTP-маршрутов API.
// Здесь связываем URL с обработчиками и навешиваем middleware (auth).
package api

import "net/http"

// Init регистрирует все маршруты API в стандартном mux.
// /api/signin — вход (выдача JWT), остальные — защищённые (auth(...)).
func Init() {
	// вход
	http.HandleFunc("/api/signin", signinHandler)

	// защищённые задачи
	http.HandleFunc("/api/task", auth(taskHandler))   // GET/POST/PUT/DELETE
	http.HandleFunc("/api/tasks", auth(tasksHandler)) // список
	http.HandleFunc("/api/task/done", auth(taskDoneHandler))

	// сервисный обработчик вычисления дат (можно тоже обернуть auth при желании)
	http.HandleFunc("/api/nextdate", nextDateHandler)
}
