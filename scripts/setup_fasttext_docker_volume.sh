#!/usr/bin/env bash
# scripts/setup_fasttext_docker_volume.sh
# ==============================================================================
# Helper script to initialize and populate the Docker Named Volume for FastText
# Pretrained Language Identification Model (lid.176.ftz).
#
# Requirements fulfilled:
# 1. Checks Docker CLI availability and Docker daemon status.
# 2. Creates Named Volume 'chatbot_fasttext_model' if not already present.
# 3. Checks if 'lid.176.ftz' already exists and is non-empty inside the volume.
# 4. Downloads model inside an isolated Alpine container via curl with retry.
# 5. Atomic write: downloads to temp file, checks size, renames to target.
# 6. Cleans up any partial download on failure.
# 7. No files are downloaded or saved to the host project directory.
# ==============================================================================

set -Eeuo pipefail

VOLUME_NAME="chatbot_fasttext_model"
MODEL_URL="https://dl.fbaipublicfiles.com/fasttext/supervised-models/lid.176.ftz"
ALPINE_IMAGE="alpine:3.21"

log_info() {
    echo "[INFO] $(date '+%Y-%m-%d %H:%M:%S') - $*"
}

log_warn() {
    echo "[WARN] $(date '+%Y-%m-%d %H:%M:%S') - $*" >&2
}

log_error() {
    echo "[ERROR] $(date '+%Y-%m-%d %H:%M:%S') - $*" >&2
}

# 1. Check Docker CLI availability
if ! command -v docker >/dev/null 2>&1; then
    log_error "Docker CLI is not installed or not in PATH."
    log_error "Please install Docker before running this script."
    exit 1
fi

# 2. Check Docker daemon status
if ! docker info >/dev/null 2>&1; then
    log_error "Docker daemon is not running or current user lacks docker permissions."
    log_error "Please start Docker Desktop / daemon and verify access with 'docker info'."
    exit 1
fi

log_info "Docker environment verified."

# 3. Create Named Volume if it does not exist
if ! docker volume inspect "${VOLUME_NAME}" >/dev/null 2>&1; then
    log_info "Creating Docker Named Volume: '${VOLUME_NAME}'..."
    docker volume create "${VOLUME_NAME}" >/dev/null
    log_info "Named Volume '${VOLUME_NAME}' created successfully."
else
    log_info "Docker Named Volume '${VOLUME_NAME}' already exists."
fi

# 4. Check if valid model already exists in the volume
log_info "Checking if model file 'lid.176.ftz' is already present in volume '${VOLUME_NAME}'..."
MODEL_EXISTS=$(docker run --rm \
    --mount type=volume,source="${VOLUME_NAME}",target=/models/language \
    "${ALPINE_IMAGE}" \
    sh -c 'if [ -s /models/language/lid.176.ftz ]; then echo "YES"; else echo "NO"; fi')

if [ "${MODEL_EXISTS}" = "YES" ]; then
    log_info "Model 'lid.176.ftz' already exists and is non-empty in volume '${VOLUME_NAME}'."
    docker run --rm \
        --mount type=volume,source="${VOLUME_NAME}",target=/models/language \
        "${ALPINE_IMAGE}" \
        ls -lh /models/language/lid.176.ftz
    log_info "Skipping download. Model volume is ready to use."
    exit 0
fi

# 5. Download model inside a temporary Alpine container using atomic write
log_info "Downloading FastText model inside temporary container '${ALPINE_IMAGE}' into volume '${VOLUME_NAME}'..."
log_info "Source URL: ${MODEL_URL}"

docker run --rm \
    --mount type=volume,source="${VOLUME_NAME}",target=/models/language \
    "${ALPINE_IMAGE}" \
    sh -ec "
        apk add --no-cache curl ca-certificates >/dev/null 2>&1

        TEMP_FILE='/models/language/lid.176.ftz.tmp'
        TARGET_FILE='/models/language/lid.176.ftz'

        trap 'rm -f \"\${TEMP_FILE}\"' EXIT INT TERM

        echo 'Starting curl download with retries...'
        curl -fL --retry 3 --connect-timeout 15 \
            -o \"\${TEMP_FILE}\" \
            '${MODEL_URL}'

        if [ ! -s \"\${TEMP_FILE}\" ]; then
            echo 'Error: Downloaded temp file is empty or missing.' >&2
            exit 1
        fi

        mv \"\${TEMP_FILE}\" \"\${TARGET_FILE}\"
        test -s \"\${TARGET_FILE}\"

        echo 'Model verification:'
        ls -lh \"\${TARGET_FILE}\"
    "

log_info "FastText model 'lid.176.ftz' successfully stored in Docker Named Volume '${VOLUME_NAME}'."
log_info "Volume setup complete."

