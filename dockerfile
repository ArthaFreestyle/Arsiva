FROM golang:1.25-alpine AS builder

RUN apk update && apk add --no-cache build-base gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
# KUNCI 1: Hapus GOARCH, dan tambahin flag -extldflags "-static" biar C library-nya dibungkus mati ke dalam binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static" -s -w' -trimpath -o server_app ./cmd/web/main.go
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server_app .
COPY --from=builder /app/docs ./docs

ENV TZ=Asia/Jakarta

EXPOSE 3000

ENTRYPOINT ["./server_app"]