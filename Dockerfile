FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/rabbitautoscaler

# Stage 2: Run
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/rabbitautoscaler .

CMD ["./rabbitautoscaler"]