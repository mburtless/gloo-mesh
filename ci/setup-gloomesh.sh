#!/bin/bash -ex

#####################################
#
# Set up gloo mesh in the target kind cluster.
#
#####################################

cluster=$1
apiServerAddress=$2

PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )/.."
source ${PROJECT_ROOT}/ci/setup-funcs.sh

if [ "${cluster}" == "" ]; then
  cluster=mgmt-cluster
fi

K="kubectl --context kind-${cluster}"

echo "deploying gloo-mesh to ${cluster} from local images..."

## build and load GlooMesh docker images
MAKE="make -C $PROJECT_ROOT"
eval "${MAKE} clean-helm manifest-gen package-helm build-all-images -B"

setChartVariables

agentCrdsChart=${AGENT_CRDS_CHART}
agentChart=${AGENT_CHART}
agentImage=${AGENT_IMAGE}
glooMeshImage=${GLOOMESH_IMAGE}
gloomeshChart=${GLOOMESH_CHART}

# Load GlooMesh discovery and networking images
# they use the same container/binary
kind load docker-image --name "${cluster}" "${glooMeshImage}"
# Load cert-agent image
kind load docker-image --name "${cluster}" "${agentImage}"

## install to kube

# set verbose to true to obtain debug logs in error dump
# set disallowIntersectingConfig for conflict detection e2e test
cat > helm-values-overrides.yaml << EOF
verbose: true
disallowIntersectingConfig: true
EOF

go run "${PROJECT_ROOT}/cmd/meshctl/main.go" install \
  --context kind-"${cluster}" \
  --chart "${gloomeshChart}" \
  --namespace gloo-mesh \
  --register \
  --cluster-name "${cluster}" \
  --verbose  \
  --api-server-address "${apiServerAddress}" \
  --agent-chart "${agentChart}" \
  --agent-crds-chart "${agentCrdsChart}" \
  --values helm-values-overrides.yaml


${K} -n gloo-mesh rollout status deployment networking
${K} -n gloo-mesh rollout status deployment discovery

echo setup successfully set up gloo-mesh
