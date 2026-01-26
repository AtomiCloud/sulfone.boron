#!/usr/bin/env bash

cluster="$1"
key="$2"

[ -z "$cluster" ] && cluster="$(printf "mica\\ntalc\\nopal\\nruby\\ntopaz\\namber\\n" | fzf)"
[ -z "$key" ] && key="$(find ~/.ssh -maxdepth 1 -name 'id_*' ! -name '*.pub' -print | fzf)"
set -euo pipefail

echo "🚪 Entering SSH session..."
json="$(tofu "-chdir=tofu/live/$cluster" output -json)"
address=$(echo "$json" | jq -r '.address.value')
ssh -i "$key" "kirin@$address"
