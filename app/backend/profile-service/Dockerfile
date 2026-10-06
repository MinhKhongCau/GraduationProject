FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy file go.mod and go.sum 
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build application come to executable file named "main"
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd

#Create lightweight image to run
FROM alpine:latest

WORKDIR /app

# Copy executable file from build stage
COPY --from=builder /app/main .

# Expose port for the application
EXPOSE 8081 9002

# Run the application
CMD ["./main"]
