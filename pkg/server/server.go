// Package server запускает HTTP-сервер приложения: регистрирует API и раздаёт статические файлы из ./web.
package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"todo/pkg/api"
)

// Start запускает простой HTTP-сервер.
// Делает три вещи:
//  1. Регистрирует API-эндпоинты (api.Init()) в стандартном mux.
//  2. Вешает раздачу статических файлов из каталога ./web на корень "/"
//     (index.html, js, css, favicon и т.п.).
//  3. Запускает http.ListenAndServe на адресе вида ":<порт>".
//
// Порт по умолчанию — 7540. Можно переопределить переменной окружения TODO_PORT.
// Пример запуска: TODO_PORT=8080 go run .
func Start() error {
	// каталог фронтенда со статикой
	webDir := "./web"

	// получаем адрес (":7540" по умолчанию или из TODO_PORT)
	addr := getAddr()

	// регистрируем API-обработчики в стандартном mux
	api.Init()

	// раздача фронтенда (тоже в стандартный mux)
	// Примеры:
	//   GET /            -> ./web/index.html
	//   GET /js/...      -> ./web/js/...
	//   GET /css/...     -> ./web/css/...
	//   GET /favicon.ico -> ./web/favicon.ico
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	// простое сообщение в консоль, чтобы видеть, что сервер поднялся
	fmt.Println("Сервер запущен на порту", addr)

	// передаём nil → используем http.DefaultServeMux, где уже всё зарегистрировано выше
	return http.ListenAndServe(addr, nil)
}

// getAddr возвращает строку адреса вида ":<порт>".
// По умолчанию используется порт 7540.
// Если переменная окружения TODO_PORT содержит число 1..65535 — берём его.
func getAddr() string {
	port := 7540
	if p := os.Getenv("TODO_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 && n <= 65535 {
			port = n
		}
	}
	return ":" + strconv.Itoa(port)
}
