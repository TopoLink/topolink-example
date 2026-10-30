#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

tool="${TOPOLINK:-go run github.com/topolink/topolink/cmd/topolink@v1}"

# The semantic fixture is read-only; work on a copy so the committed cache is
# never rewritten by a CI run.
sem="$(mktemp)"
trap 'rm -f "$sem"' EXIT
cp testdata/semantic.json "$sem"

out="$($tool -repo . -semantic "$sem" -format markdown)"

rules=(
  query-in-loop
  write-in-loop
  remote-in-transaction
  remote-under-lock
  write-without-where
  unchecked-write
  missing-timeout
  missing-auth
  money-write
  cache-no-ttl
  weak-auth
)

for r in "${rules[@]}"; do
  if ! grep -q "\`$r\`" <<<"$out"; then
    echo "missing expected rule: $r" >&2
    echo "$out" >&2
    exit 1
  fi
done

echo "ok: all ${#rules[@]} expected rules present"
