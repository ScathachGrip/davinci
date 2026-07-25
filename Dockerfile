# --- Step 1: Build Stage ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependencies manifest
COPY go.mod go.sum ./
RUN go mod download

# Copy source code files
COPY *.go ./

# Build statically compiled binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o da-vinci-bot .

# --- Step 2: Runner Stage ---
FROM alpine:latest

# Install CA certificates for outbound HTTPS API calls (GitHub & anime reaction endpoints)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy build artifact from builder stage
COPY --from=builder /app/da-vinci-bot .

# Default container port
EXPOSE 6667

# Run application
CMD ["./da-vinci-bot"]
