#!/bin/bash
set -eou xtrace

MYDIR=$(dirname $0)
CHARTDIR="${MYDIR}/../charts/cf-runtime"
VALUES_FILE="${CHARTDIR}/values.yaml"

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

# find paths to all maps having registry/repository/tag
yq -o=json '.. | select(type == "!!map" and has("registry") and has("repository") and has("tag")) | path' "$VALUES_FILE" |
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
    continue
  fi

  # write back to YAML
  echo "✍️  Writing digest back at $yq_path"
  yq -i "${yq_path}.digest = \"$digest\"" "$VALUES_FILE"
done


# yq eval-all '. as $item ireduce ({}; . * $item) | .. | select(has("image")) | path | join(".")' "$VALUES_FILE" | \
# while read -r path; do
#   registry=$(yq eval ".$path.image.registry" "$VALUES_FILE")
#   repository=$(yq eval ".$path.image.repository" "$VALUES_FILE")
#   tag=$(yq eval ".$path.image.tag" "$VALUES_FILE")

#   digest=$(get_image_digest "$registry" "$repository" "$tag")

#   if [[ -n "$digest" ]]; then
#     yq eval -i ".$path.image.digest = \"$digest\"" "$VALUES_FILE"
#   fi
# done
