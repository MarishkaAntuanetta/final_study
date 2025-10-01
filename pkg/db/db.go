// Package db инкапсулирует работу с SQLite: подключение, создание схемы
// и экспорт простого глобального соединения DB для остальных пакетов.
package db

import (
	"database/sql"
	"errors"
	"os"

	_ "modernc.org/sqlite" // SQLite-драйвер (CGO-less)
)

// DB — глобальное соединение с базой данных SQLite.
// В маленьком учебном проекте это упрощает код; в продакшене обычно
// применяют DI/контейнер и передают *sql.DB явным образом.
var DB *sql.DB

// schema — SQL-команды для первичной установки БД.
// Создаёт таблицу scheduler и индекс по date.
// Поля:
//   - id      INTEGER PRIMARY KEY AUTOINCREMENT
//   - date    CHAR(8) — дата в формате 20060102 (YYYYMMDD)
//   - title   VARCHAR(255)
//   - comment TEXT
//   - repeat  VARCHAR(128) — правило повторения (формат описан в api)
const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date CHAR(8) NOT NULL DEFAULT '',
	title VARCHAR(255) NOT NULL DEFAULT '',
	comment TEXT NOT NULL DEFAULT '',
	repeat VARCHAR(128) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

// Init открывает (или создаёт) SQLite-базу по пути dbFile,
// при первом запуске накатывает schema и сохраняет соединение в DB.
func Init(dbFile string) error {
	if dbFile == "" {
		return errors.New("empty db file path")
	}

	// Проверяем, есть ли файл базы: если нет — после открытия применим schema.
	_, statErr := os.Stat(dbFile)
	install := false
	if statErr != nil {
		install = true
	}

	// Открываем соединение через драйвер "sqlite".
	d, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	// Ping — ранняя проверка доступности/валидности соединения.
	if err := d.Ping(); err != nil {
		_ = d.Close()
		return err
	}

	// Если файл отсутствовал — создаём таблицу и индекс.
	if install {
		if _, err := d.Exec(schema); err != nil {
			_ = d.Close()
			return err
		}
	}

	// Сохраняем *sql.DB в глобальную переменную пакета.
	DB = d
	return nil
}
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
