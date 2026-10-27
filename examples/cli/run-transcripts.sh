#!/usr/bin/env bash
# run-transcripts.sh <sandbox> <outdir>
#
# Runs every examples/cli/*.session against the `burroughs` on PATH, with the
# sandbox (a built copy of examples/, plus testdata/ fetched at the tag) as the
# working directory, and writes one transcript per session to <outdir>.
#
# A session is one shell command per line. A line starting with "#cwd DIR"
# runs the following commands in <sandbox>/DIR; other "#" lines are comments.
# The transcript shows each command as "$ command", then its stdout and stderr
# (merged, in write order), then "[exit N]" when N is not 0.
set -euo pipefail
here=$(cd "$(dirname "$0")" && pwd)
sandbox=$1
outdir=$2
mkdir -p "$outdir"
for session in "$here"/*.session; do
  name=$(basename "$session" .session)
  out="$outdir/$name.txt"
  : > "$out"
  cwd=$sandbox
  while IFS= read -r line || [ -n "$line" ]; do
    case $line in
      '#cwd '*) cwd="$sandbox/${line#\#cwd }"; continue ;;
      '#'*|'') continue ;;
    esac
    printf '$ %s\n' "$line" >> "$out"
    code=0
    (cd "$cwd" && bash -c "$line") >> "$out" 2>&1 < /dev/null || code=$?
    if [ "$code" -ne 0 ]; then
      printf '[exit %d]\n' "$code" >> "$out"
    fi
  done < "$session"
done
