# Stage 1: сборка под Linux
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /flowmodel cmd/server/main.go

# Stage 2: минимальный образ
FROM alpine:latest

WORKDIR /app

COPY --from=builder /flowmodel .

RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080

CMD ["./flowmodel"]