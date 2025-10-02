FROM golang:1.23-alpine AS build

WORKDIR /app

# зависимости отдельно
COPY go.mod go.sum ./
RUN go mod download

# исходники
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o todo .

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /srv

# бинарник и фронтенд
COPY --from=build /app/todo /usr/local/bin/todo
COPY web ./web

# дефолты (можно переопределить при запуске)
ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db
# ENV TODO_PASSWORD=

# это подсказка; реальный маппинг задаётся флагом -p
EXPOSE 7540

# запуск приложения
CMD ["/usr/local/bin/todo"]
