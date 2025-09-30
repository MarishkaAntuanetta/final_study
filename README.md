# TODO Scheduler (учебный проект)

Простой веб-планировщик: фронтенд на статике (`/web`), бэкенд на Go + SQLite.

## Что умеет
- Раздача фронтенда (`/`), API:
  - `POST /api/signin` — вход по паролю (JWT в cookie `token`)
  - `GET/POST/PUT/DELETE /api/task` — получить/создать/изменить/удалить задачу
  - `GET /api/tasks` — список задач (поиск `?search=...`)
  - `POST /api/task/done?id=` — отметить выполненной (пересчитать дату или удалить)
  - `GET /api/nextdate` — расчёт следующей даты
- Аутентификация по переменной окружения `TODO_PASSWORD` (если пустая — выключена)
- Порт `TODO_PORT`, путь к БД `TODO_DBFILE`

## Задания со звёздочкой
- [x] Порт через `TODO_PORT`
- [x] Путь к БД через `TODO_DBFILE`
- [x] Поиск задач `?search=...`
- [x] Аутентификация (JWT)

## Запуск локально
```bash
go mod tidy
TODO_PASSWORD=12345 go run .
# http://localhost:7540/login.html
