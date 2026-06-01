#!/usr/bin/env bash
# Regenerate testify mocks under server/mock (kept out of server/service for Sonar coverage).
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p mock

MOCKERY=(go run github.com/vektra/mockery/v2@v2.53.4)

gen_repo_mock() {
  local name="$1"
  "${MOCKERY[@]}" \
    --name="$name" \
    --dir=./repository/mysql \
    --output=./mock \
    --outpkg=mock \
    --filename="${name}.mock.go" \
    --structname="Mock${name}" \
    --disable-version-string
}

gen_service_mock() {
  local name="$1"
  "${MOCKERY[@]}" \
    --name="$name" \
    --dir=./service \
    --output=./mock \
    --outpkg=mock \
    --filename="${name}.mock.go" \
    --structname="Mock${name}" \
    --disable-version-string
}

for name in CampaignRepository LandingPageRepository UserRepository ParticipantRepository; do
  gen_repo_mock "$name"
done

for name in AccountService; do
  gen_service_mock "$name"
done

gen_event_mock() {
  local name="$1"
  "${MOCKERY[@]}" \
    --name="$name" \
    --dir=./event \
    --output=./mock \
    --outpkg=mock \
    --filename="${name}.mock.go" \
    --structname="Mock${name}" \
    --disable-version-string
}

for name in CampaignRewardNotifier; do
  gen_event_mock "$name"
done

echo "Mocks written to server/mock/"
