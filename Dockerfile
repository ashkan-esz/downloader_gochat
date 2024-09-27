FROM golang:1.23.1-alpine as builder
WORKDIR /app

RUN apk add libwebp-dev build-base vips-dev --no-cache

COPY go.* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /app/myapp cmd/api/main.go

FROM alpine
RUN apk add libwebp-dev vips --no-cache
WORKDIR /app
COPY --from=builder /app/myapp /app/myapp

EXPOSE 7777
ENTRYPOINT [ "/app/myapp" ]
