#!/usr/bin/env bash

cluster="$1"

[ -z "$cluster" ] && cluster="$(printf "mica\ntalc\nopal\nruby\ntopaz\namber\n" | fzf)"

set -euo pipefail

echo "🚀 Deploying..."
json="$(tofu "-chdir=tofu/live/$cluster" output -json)"
address=$(echo "$json" | jq -r '.address.value')
ansible-playbook playbook.yaml -i "kirin@$address,"
echo "✅ Deployment successful"
