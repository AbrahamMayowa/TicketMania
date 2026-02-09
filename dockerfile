# -------- Build stage --------
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git (needed for private modules sometimes)
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o app \
    ./cmd/api

# -------- Runtime stage --------
FROM alpine:3.19

# Install CA certs for HTTPS
RUN apk --no-cache add ca-certificates \
    && adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/app .

# Run as non-root
USER appuser

EXPOSE 8080

ENTRYPOINT ["./app"]
