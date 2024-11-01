#!/bin/bash
set -eou xtrace

MYDIR=$(dirname $0)
CHARTDIR="${MYDIR}/../charts/cf-runtime"
VALUES_FILE="${MYDIR}/values.yaml"

runtime_images=$(yq e '.runtime.engine.runtimeImages' $CHARTDIR/values.yaml)

while read -r line; do
    key=${line%%:*}
    image=${line#*:}
    digest=$(regctl manifest digest $image)
    yq e -i ".runtime.engine.runtimeImages.$key |= . + \"$digest\"" $CHARTDIR/values.yaml
done <<< "$runtime_images"


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

yq eval-all '. as $item ireduce ({}; . * $item) | .. | select(has("image")) | path | join(".")' "$input_file" | \
while read -r path; do
  registry=$(yq eval ".$path.image.registry" "$input_file")
  repository=$(yq eval ".$path.image.repository" "$input_file")
  tag=$(yq eval ".$path.image.tag" "$input_file")

  digest=$(get_image_digest "$registry" "$repository" "$tag")

  if [[ -n "$digest" ]]; then
    yq eval -i ".$path.image.digest = \"$digest\"" "$input_file"
  fi
done
