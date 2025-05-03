#!/usr/bin/env bash

set -eou pipefail

echo "🔏 Setting up secrets for local development..."

set +e
(infisical secrets) &>/dev/null
ec="$?"
set -e

if [ "$ec" != '0' ]; then
  infisical login
fi

echo "⬇️ Downloading local secrets..."
infisical export --template=./tfvars.tpl >*.auto.tfvars
./scripts/local/sync.sh
echo "✅ Secrets set up for local development!"
