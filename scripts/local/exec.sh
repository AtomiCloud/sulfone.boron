#!/usr/bin/env bash

cluster="$1"

[ -z "$cluster" ] && cluster="$(printf "mica\ntalc\nopal\nruby\ntopaz\namber\n" | fzf)"

set -euo pipefail

echo "🚪 Entering SSH session..."
json="$(tofu "-chdir=tofu/live/$cluster" output -json)"
address=$(echo "$json" | jq -r '.address.value')
ssh "kirin@$address"
