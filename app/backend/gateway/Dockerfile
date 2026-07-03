# ---- Build Go Plugin ----
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy all source code first
COPY go-plugins/mindcare-auth/ .

# Run go mod tidy to resolve dependencies based on main.go
RUN go mod tidy

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /mindcare-auth main.go

# ---- Gateway runtime stage ----
FROM kong:alpine

USER root

# Copy plugin binary
COPY --from=builder /mindcare-auth /usr/local/bin/mindcare-auth

# Copy declarative config (routes/plugins)
COPY kong.yml /usr/local/kong/declarative/kong.yml

# Cấu hình biến môi trường để Kong kích hoạt Go Plugin
ENV KONG_PLUGINS=bundled,mindcare-auth
ENV KONG_PLUGINSERVER_NAMES=mindcare-auth
ENV KONG_PLUGINSERVER_MINDCARE_AUTH_QUERY_CMD="/usr/local/bin/mindcare-auth -dump"
ENV KONG_PLUGINSERVER_MINDCARE_AUTH_START_CMD="/usr/local/bin/mindcare-auth"

USER kong
