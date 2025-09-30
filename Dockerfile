# сборка бинарника
FROM golang:1.22 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o todo .

# минимальный рантайм
FROM ubuntu:latest
WORKDIR /srv
COPY --from=build /app/todo /usr/local/bin/todo
COPY web ./web
EXPOSE 7540
# переменные (можно переопределять при запуске)
ENV TODO_PORT=7540
# ENV TODO_DBFILE=/data/scheduler.db
# ENV TODO_PASSWORD=

# том с БД на хосте будем монтировать в /data (см. пример запуска ниже)
CMD ["todo"]
