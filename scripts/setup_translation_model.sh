#!/usr/bin/env bash
# -----------------------------------------------------------------------------
# Setup M2M100 418M Translation Model into Docker Named Volume
#
# Volume: chatbot_translation_model
# Mount:  /models/translation
# Target: /models/translation/m2m100_418M
# Repo:   facebook/m2m100_418M (revision: 55c2e61bbf05dfb8d7abccdc3fae6fc8512fd636)
# -----------------------------------------------------------------------------
set -euo pipefail

# 1. Verify Docker CLI
if ! command -v docker >/dev/null 2>&1; then
    echo "❌ Error: Docker CLI is not installed or not in PATH." >&2
    exit 1
fi

# 2. Verify Docker daemon
if ! docker info >/dev/null 2>&1; then
    echo "❌ Error: Docker daemon is not running or unreachable." >&2
    exit 1
fi

# 3. Inspect or create named volume (never delete or recreate)
VOLUME_NAME="chatbot_translation_model"
if ! docker volume inspect "$VOLUME_NAME" >/dev/null 2>&1; then
    echo "Creating Docker volume: $VOLUME_NAME"
    docker volume create "$VOLUME_NAME" >/dev/null
else
    echo "Docker volume '$VOLUME_NAME' already exists."
fi

# 4. Run temporary python:3.10-slim container to download, validate, and publish
echo "Starting temporary Python 3.10 container with volume mount..."

docker run --rm \
    -v "${VOLUME_NAME}:/models/translation" \
    python:3.10-slim \
    bash -c '
set -euo pipefail

# Install pinned huggingface_hub inside the container
pip install --no-cache-dir "huggingface_hub==0.21.4" >/dev/null

python3 - << "PYEOF"
import os
import shutil
import sys
from pathlib import Path
from huggingface_hub import snapshot_download

REPO_ID = "facebook/m2m100_418M"
REVISION = "55c2e61bbf05dfb8d7abccdc3fae6fc8512fd636"

MOUNT_BASE = Path("/models/translation")
FINAL_DIR = MOUNT_BASE / "m2m100_418M"
STAGING_DIR = MOUNT_BASE / ".m2m100_418M.tmp"

REQUIRED_FILES = [
    "config.json",
    "generation_config.json",
    "pytorch_model.bin",
    "sentencepiece.bpe.model",
    "special_tokens_map.json",
    "tokenizer_config.json",
    "vocab.json",
]

def is_complete_model(path: Path) -> bool:
    if not path.is_dir():
        return False
    for filename in REQUIRED_FILES:
        filepath = path / filename
        # Must not be a symbolic link (broken or valid)
        if filepath.is_symlink():
            return False
        # Must be a regular file
        if not filepath.is_file():
            return False
        # Must have size > 0
        try:
            if filepath.stat().st_size <= 0:
                return False
        except OSError:
            return False
    return True

# 1. Check if complete model already exists
if FINAL_DIR.exists():
    if is_complete_model(FINAL_DIR):
        print(f"✅ Translation model is complete and ready at: {FINAL_DIR}")
        sys.exit(0)
    else:
        print(f"⚠️ Model directory {FINAL_DIR} exists but is incomplete or invalid (contains broken symlinks or missing files).")
        print(f"Removing invalid directory: {FINAL_DIR}")
        if FINAL_DIR.is_dir() and not FINAL_DIR.is_symlink():
            shutil.rmtree(FINAL_DIR, ignore_errors=True)
        else:
            FINAL_DIR.unlink(missing_ok=True)

# 2. Clean up any leftover staging directory from a previous interrupted run
if STAGING_DIR.exists():
    print(f"Cleaning leftover staging directory: {STAGING_DIR}")
    if STAGING_DIR.is_dir() and not STAGING_DIR.is_symlink():
        shutil.rmtree(STAGING_DIR, ignore_errors=True)
    else:
        STAGING_DIR.unlink(missing_ok=True)

STAGING_DIR.mkdir(parents=True, exist_ok=True)

published = False
try:
    print(f"Downloading {REPO_ID} (rev: {REVISION}) into staging: {STAGING_DIR}...")
    snapshot_download(
        repo_id=REPO_ID,
        revision=REVISION,
        local_dir=str(STAGING_DIR),
        local_dir_use_symlinks=False,
        allow_patterns=REQUIRED_FILES,
        local_files_only=False,
        resume_download=True,
    )

    # 3. Validate that every required file exists, is regular, non-symlink, and non-empty
    if not is_complete_model(STAGING_DIR):
        raise RuntimeError("Validation failed: Not all required files were downloaded, some are empty, or files are symlinks.")

    # 4. Atomic publication via same-filesystem directory rename
    if FINAL_DIR.exists():
        raise RuntimeError(f"Target directory {FINAL_DIR} unexpectedly appeared before publication.")

    os.rename(str(STAGING_DIR), str(FINAL_DIR))
    published = True
    print(f"✅ Translation model successfully published to: {FINAL_DIR}")

except Exception as err:
    print(f"❌ Error during translation model setup: {err}", file=sys.stderr)
    sys.exit(1)
finally:
    # Guarantee immediate cleanup of staging if download or validation failed
    if not published and STAGING_DIR.exists():
        print(f"Cleaning incomplete staging directory: {STAGING_DIR}", file=sys.stderr)
        if STAGING_DIR.is_dir() and not STAGING_DIR.is_symlink():
            shutil.rmtree(STAGING_DIR, ignore_errors=True)
        else:
            STAGING_DIR.unlink(missing_ok=True)

PYEOF
'
