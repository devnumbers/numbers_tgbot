FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o tg-bot .

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/tg-bot .

EXPOSE 3002

CMD ["./tg-bot"]