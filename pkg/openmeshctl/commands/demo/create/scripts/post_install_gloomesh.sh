#!/bin/bash -ex

cluster=$0

K="kubectl --context kind-${cluster}"

# sleep to allow namespace to be created
sleep 2

echo "Running post-install checks..."

${K} -n gloo-mesh rollout status deployment networking
${K} -n gloo-mesh rollout status deployment discovery

# sleep to allow CRDs to register
sleep 4

