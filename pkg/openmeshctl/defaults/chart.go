package defaults

import (
	"fmt"
	"strings"
)

const (
	// Namespace is the namespace used to install charts to.
	Namespace = "gloo-mesh"

	// ReleaseName is the default name of the Gloo Mesh helm installation.
	ReleaseName = "gloo-mesh"

	// ClusterDomain is the default domain of a cluster for local traffic.
	ClusterDomain = "cluster.local"

	// AgentReleaseName is the default name of the Gloo Mesh Cert Agent
	// installation.
	AgentReleaseName = "cert-agent"

	// AgentCRDsReleaseName is the default name of the Gloo Mesh Cert Agent CRDs
	// installation.
	AgentCRDsReleaseName = "agent-crds"
)

const repoURI = "https://storage.googleapis.com/gloo-mesh"

// GlooMeshChartURI returns the URI to Gloo Mesh chart for the given version.
func GlooMeshChartURI(version string) string {
	return fmt.Sprintf("%s/gloo-mesh/gloo-mesh-%s.tgz", repoURI, sanitizeVersion(version))
}

// AgentCRDsChartURI returns the URI to the cert agent CRDs chart for the given version.
func AgentCRDsChartURI(version string) string {
	return fmt.Sprintf("%s/agent-crds/agent-crds-%s.tgz", repoURI, sanitizeVersion(version))
}

// CertAgentChartURI returns the URI to the cert agent chart for the given version.
func CertAgentChartURI(version string) string {
	return fmt.Sprintf("%s/cert-agent/cert-agent-%s.tgz", repoURI, sanitizeVersion(version))
}

func sanitizeVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}
