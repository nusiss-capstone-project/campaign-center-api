#!/usr/bin/env bash
# Deprecated: use scripts/gen-mocks.sh (mocks live under server/mock, not server/service/mock).
set -euo pipefail
exec "$(dirname "$0")/gen-mocks.sh" "$@"
