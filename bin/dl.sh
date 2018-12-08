#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

url=$(grep -E -o -m1 'https?://adventofcode.com/[0-9]*/day/[0-9]*')
curl -s --cookie cookie.txt "$url" \
| pandoc -f html -t markdown
