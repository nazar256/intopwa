#!/bin/bash

# List of KV namespace IDs
KV_NAMESPACE_IDS=("38744dc4fa434387874678f549c9cd91" "97c361cb39c944d4bc466433da6af203")

# Iterate over each namespace ID
for KV_NAMESPACE_ID in "${KV_NAMESPACE_IDS[@]}"; do
  echo "Processing namespace: $KV_NAMESPACE_ID"

  # List all keys in the KV namespace
  keys=$(npx wrangler kv:key list --namespace-id $KV_NAMESPACE_ID | jq -r '.[].name')

  # Iterate over each key and delete it
  for key in $keys; do
    npx wrangler kv:key delete --namespace-id $KV_NAMESPACE_ID $key
  done

  echo "All keys have been deleted from namespace: $KV_NAMESPACE_ID"
done