# ---- Build ----
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-service ./cmd

# ---- Runtime ----
FROM alpine:3.20 AS runner
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/payment-service .

EXPOSE 8082
ENTRYPOINT ["./payment-service"]
