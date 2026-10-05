# ---- Build ----
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -p 2 -o booking-service ./cmd

# ---- Runtime ----
FROM alpine:3.20 AS runner
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/booking-service .

EXPOSE 8083
ENTRYPOINT ["./booking-service"]
