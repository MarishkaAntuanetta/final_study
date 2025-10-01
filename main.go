// Package main — точка входа приложения-планировщика.
// Здесь инициализируется база данных и запускается веб-сервер.
package main

import (
	"log"
	"os"

	"todo/pkg/db"
	"todo/pkg/server"
)

// main — старт программы.
// 1) Определяем путь к файлу БД (переменная окружения TODO_DBFILE или "scheduler.db").
// 2) Инициализируем SQLite (создаём файл и таблицы при первом запуске).
// 3) Запускаем веб-сервер (порт по умолчанию 7540, можно задать TODO_PORT).
// Если на любом шаге произойдёт ошибка — выводим её и выходим.
func main() {
	// файл базы данных по умолчанию
	dbFile := "scheduler.db"

	// звёздочка: разрешаем задавать путь к БД извне через переменную окружения
	// Пример: TODO_DBFILE=/data/scheduler.db go run .
	if env := os.Getenv("TODO_DBFILE"); env != "" {
		dbFile = env
	}

	// инициализация БД:
	// - открывает соединение к SQLite
	// - если файла не было — создаёт таблицу scheduler и индекс по date
	if err := db.Init(dbFile); err != nil {
		log.Fatal(err) // критическая ошибка — завершаем программу
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("db close error: %v\n", err)
		}
	}()
	// запуск HTTP-сервера:
	// - раздаёт статические файлы из ./web (index.html, css, js, favicon)
	// - регистрирует API: /api/signin, /api/task, /api/tasks, /api/task/done, /api/nextdate
	// - порт по умолчанию :7540, можно переопределить переменной TODO_PORT
	// - если указана TODO_PASSWORD — включается простая аутентификация (JWT в cookie "token")
	if err := server.Start(); err != nil {
		log.Printf("server error: %v\n", err) // если сервер упал — логируем и выходим
	}
}
