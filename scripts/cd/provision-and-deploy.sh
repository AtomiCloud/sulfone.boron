#!/usr/bin/env bash

cluster="$1"
key="$2"

set -euo pipefail

# deploy compute here
[ -z "$cluster" ] && cluster="$(printf "mica\ntalc\nopal\nruby\ntopaz\namber\n" | fzf)"
echo "🚀 Provisioning to $cluster..."

echo "🆕 Initializing $cluster..."
tofu "-chdir=tofu/live/$cluster" init
echo "✅ $cluster initialized successfully"

echo "🚀 Provisioning $cluster..."
tofu "-chdir=tofu/live/$cluster" apply
echo "✅ $cluster deployed successfully"

echo "📤 Extracting IP address..."
json="$(tofu "-chdir=tofu/live/$cluster" output -json)"
address=$(echo "$json" | jq -r '.address.value')
echo "✅ IP address extracted successfully: $address"

echo "📝 Writing IP address to general/ip.tfvars..."
echo "ip_address = \"$address\"" >./tofu/live/general/ip.tfvars
echo "✅ IP address written to general/ip.tfvars successfully"

echo "🆕 Initializing Generic Infrastructure..."
tofu -chdir=tofu/live/general init
echo "✅ Generic Infrastructure initialized successfully"

echo "🏗️ Provisioning Generic Infrastructure..."
tofu -chdir=tofu/live/general apply -var-file="ip.tfvars"
echo "✅ Generic Infrastructure provisioned successfully"

echo "🚀 Deploying..."
./scripts/cd/deploy.sh "$cluster" "$key"
echo "✅ Deployment successful"
