#!/usr/bin/env bash
# Regenerate mockery mocks under server/service/mock (import path .../server/service/mock).
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p service/mock
for name in CampaignRepository LandingPageRepository UserRepository ParticipantRepository RewardTransactionRepository; do
  go run github.com/vektra/mockery/v2@v2.53.4 \
    --name="$name" \
    --dir=./repository/mysql \
    --output=./service/mock \
    --outpkg=mock \
    --filename="${name}.mock.go" \
    --structname="Mock${name}" \
    --disable-version-string
done
