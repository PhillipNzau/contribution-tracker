# ---- Build stage ----
FROM golang:1.24 AS builder

WORKDIR /app

# Copy go.mod and go.sum from backend folder
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the rest of the backend source code
COPY backend/ .

# Build the Go binary
RUN go build -o server main.go

# ---- Run stage ----
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]
