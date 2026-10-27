#!/usr/bin/env bash
# run-examples.sh [--update]
#
# Runs every example under examples/ against a Burroughs release, from the Go
# module proxy, and compares what each prints with its committed expected
# output. Any mismatch, build error or non-zero exit fails the run.
#
#   BURROUGHS_TAG   the release to test against; default: the latest vX.Y.Z tag,
#                   resolved through the GitHub API (scripts/latest_tag.py).
#   WORK            scratch directory; default: a new temporary directory.
#   --update        write the outputs back over the committed expected files
#                   (for when a new release legitimately changes them; review
#                   the diff, and the docs that include those files, before
#                   committing).
#
# The committed go.mod files require v0.11.0 with no local replace. This script
# copies examples/ to scratch and points each copy's require at the tag, so the
# committed tree is never modified (except by --update, and then only the
# expected-output files).
set -euo pipefail

root=$(cd "$(dirname "$0")/.." && pwd)
update=0
if [ "${1:-}" = "--update" ]; then update=1; fi

py=${PYTHON:-python3}
tag=${BURROUGHS_TAG:-$("$py" -I "$root/scripts/latest_tag.py")}
work=${WORK:-$(mktemp -d)}
mkdir -p "$work"
ex="$work/examples"
rm -rf "$ex" "$work/bin" "$work/transcripts"
cp -R "$root/examples" "$ex"

echo "run-examples: burroughs $tag"
echo "run-examples: $(go version)"
echo "run-examples: $(uname -sm)"
echo "run-examples: scratch $work"

export GOFLAGS=-mod=mod
export GOWORK=off

# 1. Point every host module at the tag. (The wasip1 guest has no dependency.)
for mod in embed-core refusals component describe wasip1/host; do
  (cd "$ex/$mod" && go mod edit -require="github.com/scttfrdmn/burroughs@$tag" && go mod tidy)
done

# 2. Fetch the component fixtures from the engine repository at the tag.
mkdir -p "$ex/testdata" "$ex/component/testdata"
while IFS= read -r f || [ -n "$f" ]; do
  case $f in '#'*|'') continue ;; esac
  curl -fsSL --retry 3 \
    "https://raw.githubusercontent.com/scttfrdmn/burroughs/$tag/internal/component/testdata/$f" \
    -o "$ex/testdata/$f"
  cp "$ex/testdata/$f" "$ex/component/testdata/$f"
done < "$root/examples/fixtures.txt"

# 3. Build the wasip1 guest with the stock toolchain, and the CLI at the tag.
(cd "$ex/wasip1/guest" && GOOS=wasip1 GOARCH=wasm go build -o ../guest.wasm .)
mkdir -p "$work/bin"
(cd "$ex/embed-core" && go build -o "$work/bin/burroughs" github.com/scttfrdmn/burroughs/cmd/burroughs)
export PATH="$work/bin:$PATH"

fail=0
# check NAME EXPECTED ACTUAL
check() {
  if [ "$update" = 1 ]; then
    cp "$3" "$root/$2"
    echo "updated $2"
  elif diff -u "$root/$2" "$3"; then
    echo "ok      $1"
  else
    echo "FAIL    $1 (output differs from $2)"
    fail=1
  fi
}

# run NAME DIR EXPECTED [ENV...]: `go run .` in DIR must exit 0 and print EXPECTED.
run() {
  local name=$1 dir=$2 expected=$3
  shift 3
  local out="$work/out-$name.txt"
  if (cd "$ex/$dir" && env "$@" go run . > "$out"); then
    check "$name" "$expected" "$out"
  else
    echo "FAIL    $name (go run exited non-zero)"
    fail=1
  fi
}

run embed-core       embed-core  examples/embed-core/output.txt
run refusals         refusals    examples/refusals/output.txt
run refusals-threads0 refusals   examples/refusals/output-threads0.txt BURROUGHS_THREADS=0
run wasip1-host      wasip1/host examples/wasip1/output.txt
run component        component   examples/component/output.txt
run describe         describe    examples/describe/output.txt

# 4. The CLI transcripts the docs show.
"$ex/cli/run-transcripts.sh" "$ex" "$work/transcripts"
for t in "$work"/transcripts/*.txt; do
  name=$(basename "$t" .txt)
  check "cli/$name" "examples/cli/transcripts/$name.txt" "$t"
done
# Every committed transcript must still have a session.
for t in "$root"/examples/cli/transcripts/*.txt; do
  [ -e "$t" ] || continue
  name=$(basename "$t" .txt)
  if [ ! -e "$root/examples/cli/$name.session" ]; then
    echo "FAIL    cli/$name (transcript has no session)"
    fail=1
  fi
done

if [ "$fail" != 0 ]; then
  echo "run-examples: FAILED against burroughs $tag"
  exit 1
fi
echo "run-examples: all examples match against burroughs $tag"
