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

    # profile/v1/profile.proto: profile-service (server) + booking-service (client)
    for svc in profile-service booking-service; do
      out=\"\$svc/internal/infrastructure/grpc/profilepb\"
      mkdir -p \"\$out\"
      protoc -I proto \
        --go_out=\"\$out\" --go_opt=paths=source_relative \
        --go_opt=Mprofile/v1/profile.proto=\$svc/internal/infrastructure/grpc/profilepb \
        --go-grpc_out=\"\$out\" --go-grpc_opt=paths=source_relative \
        --go-grpc_opt=Mprofile/v1/profile.proto=\$svc/internal/infrastructure/grpc/profilepb \
        profile/v1/profile.proto
      mv \"\$out\"/profile/v1/*.go \"\$out\"/
      rm -rf \"\$out\"/profile
      chown -R ${HOST_UID}:${HOST_GID} \"\$svc/internal/infrastructure/grpc\"
    done
  "
