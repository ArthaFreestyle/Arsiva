FROM golang:1.25-alpine AS builder

RUN apk update && apk add --no-cache build-base gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o server_app ./cmd/web/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/server_app /server_app

ENV TZ=Asia/Jakarta

EXPOSE 3000

ENTRYPOINT ["/server_app"]