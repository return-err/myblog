FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \ 
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/myblog ./cmd/server

FROM alpine:3.24

RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

COPY --from=builder --chown=app:app /out/myblog .
COPY --from=builder --chown=app:app /app/templates ./templates
COPY --from=builder --chown=app:app /app/static ./static

USER app

EXPOSE 8080

CMD ["./myblog"]
