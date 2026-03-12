#!/usr/bin/env bash
set -euo pipefail

FAILED=0

check_empty() {
  local output="$1"
  local message="$2"

  if [[ -n "$output" ]]; then
    echo "FAIL: $message"
    echo "$output"
    FAILED=1
  fi
}

svc_to_svc=$(rg -n '"github.com/MikeS071/archonhq/services/' services --glob '*.go' || true)
check_empty "$svc_to_svc" "service-to-service direct imports are not allowed"

pkg_bad=$(rg -n '"github.com/MikeS071/archonhq/(apps|services|integrations)/' pkg --glob '*.go' || true)
check_empty "$pkg_bad" "pkg must only import pkg dependencies"

integration_bad=$(rg -n '"github.com/MikeS071/archonhq/(apps|services)/' integrations --glob '*.go' || true)
check_empty "$integration_bad" "integrations may only depend on pkg"

if [[ "$FAILED" -eq 1 ]]; then
  exit 1
fi

echo "PASS: import boundaries are clean"
