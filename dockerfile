FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o server_app ./cmd/web/main.go

FROM alpine:latest

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Asia/Jakarta

COPY --from=builder /app/server_app /server_app

EXPOSE 3000

ENTRYPOINT ["/server_app"]