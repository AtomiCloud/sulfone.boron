#!/usr/bin/env bash

cluster="$1"
key="$2"

[ -z "$cluster" ] && cluster="$(printf "mica\ntalc\nopal\nruby\ntopaz\namber\n" | fzf)"
[ -z "$key" ] && key="$(find ~/.ssh -maxdepth 1 -name 'id_*' ! -name '*.pub' -print | fzf)"

set -euo pipefail

echo "🏷️  Getting latest git tag..."
VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -z "$VERSION" ]; then
  echo "⚠️  No git tag found, falling back to 'latest'"
  VERSION="latest"
else
  echo "✅ Using version: $VERSION"
fi

echo "🚀 Deploying..."
json="$(tofu "-chdir=tofu/live/$cluster" output -json)"
address=$(echo "$json" | jq -r '.address.value')
ansible-playbook playbook.yaml -i "kirin@$address," --private-key "$key" -e "boron_version=$VERSION"
echo "✅ Deployment successful"
