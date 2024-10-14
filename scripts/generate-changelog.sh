#!/bin/bash

# Check if the correct number of arguments is provided
if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <path_to_Chart.yaml> <path_to_CHANGELOG.md>"
  exit 1
fi

# Assign input arguments to variables
CHART_YAML="$1"
CHANGELOG_FILE="$2"

# Check if the Chart.yaml file exists
if [ ! -f "$CHART_YAML" ]; then
  echo "Error: Chart.yaml file not found at $CHART_YAML"
  exit 1
fi

# Extract the artifacthub.io/changes section from the Chart.yaml
CHANGES=$(sed -n '/artifacthub.io\/changes: |/,/^dependencies:/p' "$CHART_YAML" | sed '1d;$d')

echo $CHANGES

# Create associative arrays to store changes by kind
declare -A changes_by_kind

# Iterate through the changes and group them by kind
current_kind=""
while read -r line; do
  if [[ $line == *"- kind:"* ]]; then
    current_kind=$(echo "$line" | sed 's/.*kind: //')
    # Initialize an empty array for the kind if it doesn't exist
    if [[ -z "${changes_by_kind[$current_kind]}" ]]; then
      changes_by_kind[$current_kind]=""
    fi
  elif [[ $line == *"description:"* ]]; then
    description=$(echo "$line" | sed 's/.*description: "//;s/"//')
    # Append the description to the corresponding kind
    changes_by_kind[$current_kind]+="- $description"$'\n'
  fi
done <<< "$CHANGES"

# Create the CHANGELOG.md file and write the header
echo "# Changelog" > "$CHANGELOG_FILE"
echo "" >> "$CHANGELOG_FILE"

# Write the changes grouped by kind to the CHANGELOG.md file
for kind in "${!changes_by_kind[@]}"; do
  if [[ -n "${changes_by_kind[$kind]}" ]]; then
    echo "## $kind" >> "$CHANGELOG_FILE"
    echo "${changes_by_kind[$kind]}" >> "$CHANGELOG_FILE"
  fi
done
