FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o myblog ./cmd/server/main.go


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/myblog .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./myblog"]
