FROM golang:1.21-alpine as builder
WORKDIR /app

RUN apk add libwebp-dev build-base vips vips-dev --no-cache \
    --repository https://dl-3.alpinelinux.org/alpine/edge/testing/ \
    --repository https://dl-3.alpinelinux.org/alpine/edge/main

COPY go.* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /app/myapp cmd/api/main.go

FROM alpine
RUN apk add libwebp-dev vips --no-cache \
    --repository https://dl-3.alpinelinux.org/alpine/edge/testing/ \
    --repository https://dl-3.alpinelinux.org/alpine/edge/main
WORKDIR /app
COPY --from=builder /app/myapp /app/myapp

EXPOSE 7777
ENTRYPOINT [ "/app/myapp" ]
