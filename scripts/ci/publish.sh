#!/usr/bin/env bash

[ "${DOMAIN}" = '' ] && echo "❌ 'DOMAIN' env var not set" && exit 1

[ "${DOCKER_PASSWORD}" = '' ] && echo "❌ 'DOCKER_PASSWORD' env var not set" && exit 1
[ "${DOCKER_USER}" = '' ] && echo "❌ 'DOCKER_USER' env var not set" && exit 1

[ "${GITHUB_SHA}" = '' ] && echo "❌ 'GITHUB_SHA' env var not set" && exit 1
[ "${GITHUB_BRANCH}" = '' ] && echo "❌ 'GITHUB_BRANCH' env var not set" && exit 1
[ "${GITHUB_REPO_REF}" = '' ] && echo "❌ 'GITHUB_REPO_REF' env var not set" && exit 1
[ "${CI_DOCKER_IMAGES}" = '' ] && echo "❌ 'CI_DOCKER_IMAGES' env var not set" && exit 1

version="$1"

set -euo pipefail

echo "🔐 Logging into docker registry..."
echo "${DOCKER_PASSWORD}" | docker login "${DOMAIN}" -u "${DOCKER_USER}" --password-stdin
echo "✅ Successfully logged into docker registry!"

OLDIFS=$IFS
IFS=','

for docker_image in ${CI_DOCKER_IMAGES}; do

  echo "📝 Generating Image tags for '${docker_image}' ..."
  # Obtain image
  IMAGE_ID="${DOMAIN}/${GITHUB_REPO_REF}/${docker_image//[._-]*$//}"
  IMAGE_ID=$(echo "${IMAGE_ID}" | tr '[:upper:]' '[:lower:]') # convert to lower case

  # obtaining the version
  SHA="$(echo "${GITHUB_SHA}" | head -c 6)"
  BRANCH="${GITHUB_BRANCH//[._-]*$//}"
  IMAGE_VERSION="${SHA}-${BRANCH}"

  # Generate image references
  IMAGE_REF="${IMAGE_ID}:${IMAGE_VERSION}"
  TARGET_IMAGE_REF="${IMAGE_ID}:${version}"

  # pull older image
  echo "📥 Pulling image '${IMAGE_REF}'..."
  docker pull "${IMAGE_REF}"
  echo "✅ Successfully pulled image '${IMAGE_REF}'!"

  echo "🏷️ Tagging image '${IMAGE_REF}' as '${TARGET_IMAGE_REF}'..."
  docker tag "${IMAGE_REF}" "${TARGET_IMAGE_REF}"
  echo "✅ Successfully tagged image '${IMAGE_REF}' as '${TARGET_IMAGE_REF}'!"

  echo "📤 Pushing image '${TARGET_IMAGE_REF}'..."
  docker push "${TARGET_IMAGE_REF}"
  echo "✅ Successfully pushed image '${TARGET_IMAGE_REF}'!"

done
IFS=$OLDIFS

echo "✅ Successfully published images!"
