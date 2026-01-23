#!/usr/bin/env bash

cluster="$1"
key="$2"

[ -z "$cluster" ] && cluster="$(printf "mica\ntalc\nopal\nruby\ntopaz\namber\n" | fzf)"
[
[ -z "$key" ] && key="$(ls ~/.ssh/id* | grep -v '.pub' | fzf)"
set -euo pipefail

echo "🚪 Entering SSH session..."
json="$(tofu "-chdir=tofu/live/$cluster" output -json)"
address=$(echo "$json" | jq -r '.address.value')
ssh -i "$key" "kirin@$address"
