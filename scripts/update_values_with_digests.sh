#!/bin/bash
set -eou xtrace

MYDIR=$(dirname $0)
CHARTDIR="${MYDIR}/../charts/cf-runtime"

# $1: target file (relative to CHARTDIR) to update in place. Default: values.yaml
# $2: base file (relative to CHARTDIR) used only as a fallback source for
#     registry/repository when the target is a partial overlay (e.g. rootless
#     values files that only override tag/digest). Default: same as target.
TARGET_FILE_NAME="${1:-values.yaml}"
BASE_FILE_NAME="${2:-$TARGET_FILE_NAME}"

VALUES_FILE="${CHARTDIR}/${TARGET_FILE_NAME}"
BASE_FILE="${CHARTDIR}/${BASE_FILE_NAME}"

get_image_digest() {
  local registry=$1
  local repository=$2
  local tag=$3

  digest=$(regctl manifest digest "${registry}/${repository}:${tag}" 2>/dev/null)

  if [[ $? -ne 0 ]]; then
    echo "Failed to get digest for ${registry}/${repository}:${tag}"
    echo ""
  else
    echo "$digest"
  fi
}

# find paths to all maps having tag/digest (registry/repository may live only
# in the base file, e.g. for rootless overlay files)
yq -o=json '.. | select(type == "!!map" and has("tag") and has("digest")) | path' "$VALUES_FILE" |
jq -c '.' |
while IFS= read -r path_json; do
  # build yq path expression
  yq_path=""
  for key in $(echo "$path_json" | jq -r '.[]'); do
    if [[ "$key" =~ ^[0-9]+$ ]]; then
      yq_path+="[$key]"
    else
      yq_path+=".$key"
    fi
  done

  # extract registry/repo/tag at this path
  registry=$(yq -r "${yq_path}.registry" "$VALUES_FILE")
  repository=$(yq -r "${yq_path}.repository" "$VALUES_FILE")
  tag=$(yq -r "${yq_path}.tag" "$VALUES_FILE")

  # fall back to the base file for registry/repository when the target is a
  # partial overlay (e.g. rootless values files)
  if [[ -z "$registry" || "$registry" == "null" ]]; then
    registry=$(yq -r "${yq_path}.registry" "$BASE_FILE")
  fi
  if [[ -z "$repository" || "$repository" == "null" ]]; then
    repository=$(yq -r "${yq_path}.repository" "$BASE_FILE")
  fi

  # skip if any are missing
  if [[ -z "$registry" || -z "$repository" || -z "$tag" || "$registry" == "null" || "$repository" == "null" || "$tag" == "null" ]]; then
    echo "⚠️  Skipping incomplete entry at $yq_path"
    continue
  fi

  image="${registry}/${repository}:${tag}"
  echo "🔎 Checking image: $image"

  if digest=$(regctl image digest "$image" 2>/dev/null); then
    echo "✅ Digest: $digest"
  else
    echo "❌ Failed to get digest for $image"
    exit 1
  fi

  # write back to YAML
  echo "✍️  Writing digest back at $yq_path"
  yq -i "${yq_path}.digest = \"$digest\"" "$VALUES_FILE"
done
