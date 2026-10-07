#!/usr/bin/env sh
# Sinh code Go cho các file .proto trong app/backend/proto vào từng service dùng chúng.
# Chạy trong container golang nên máy dev không cần cài protoc/Go:
#   ./scripts/gen-proto.sh
set -eu

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PROTOC_GEN_GO_VERSION="v1.36.11"
PROTOC_GEN_GO_GRPC_VERSION="v1.5.1"

HOST_UID="$(id -u)"
HOST_GID="$(id -g)"

docker run --rm \
  -e HOME=/tmp -e GOPATH=/tmp/go -e GOCACHE=/tmp/go-cache \
  -v "$ROOT/app/backend:/backend" \
  -w /backend \
  golang:1.26-alpine sh -euc "
    apk add --no-cache protobuf protobuf-dev >/dev/null
    go install google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}
    export PATH=\$PATH:/tmp/go/bin

    # gen <proto file> <go package dir name> <service...>
    gen() {
      proto=\"\$1\"; pkg=\"\$2\"; shift 2
      for svc in \"\$@\"; do
        out=\"\$svc/internal/infrastructure/grpc/\$pkg\"
        mkdir -p \"\$out\"
        protoc -I proto \
          --go_out=\"\$out\" --go_opt=paths=source_relative \
          --go_opt=M\$proto=\$svc/internal/infrastructure/grpc/\$pkg \
          --go-grpc_out=\"\$out\" --go-grpc_opt=paths=source_relative \
          --go-grpc_opt=M\$proto=\$svc/internal/infrastructure/grpc/\$pkg \
          \"\$proto\"
        mv \"\$out\"/\$(dirname \"\$proto\")/*.go \"\$out\"/
        rm -rf \"\$out\"/\$(echo \"\$proto\" | cut -d/ -f1)
        chown -R ${HOST_UID}:${HOST_GID} \"\$svc/internal/infrastructure/grpc\"
      done
    }

    # profile/v1/profile.proto: profile-service (server) + booking-service (client)
    gen profile/v1/profile.proto profilepb profile-service booking-service
    # booking/v1/booking_payment.proto: booking-service (server) + payment-service (client)
    gen booking/v1/booking_payment.proto bookingpb booking-service payment-service
    # payment/v1/payment_events.proto: payment-service (publisher) + booking-service (consumer)
    gen payment/v1/payment_events.proto paymentpb payment-service booking-service
  "
