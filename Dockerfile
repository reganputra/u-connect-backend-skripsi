FROM golang:1.23-alpine AS builder
WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /server ./main.go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /server /server

# Expose the default port (can be overridden by env var at runtime)
EXPOSE 8080

# Recommended runtime env: let the platform provide `PORT` or `APP_PORT`.
ENV APP_PORT=8080

CMD ["/server"]
