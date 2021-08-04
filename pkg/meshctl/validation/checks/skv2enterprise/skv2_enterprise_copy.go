package skv2enterprise

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"

	"github.com/rotisserie/eris"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/go-utils/grpcutils"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO import this logic from skv2-enterprise once this check is consolidated into GME
// Logic copied from skv2-enterprise to avoid a direct dependency from OSS to closed source repo

// Copied from https://github.com/solo-io/skv2-enterprise/blob/1bcee7a74da086a121e876c9f1f7e9ff83ab0f90/relay/pkg/agent/agent_setup.go#L21-L61
type Options struct {
	Server      grpcutils.DialOpts
	Cluster     string
	RelayHost   string
	AgentLabels map[string]string

	// Reference to a Secret containing the Client TLS Certificates used to identify the Relay Agent to the Server.
	// If the secret does not exist, a Token and Root cert secret are required.
	ClientCertSecret client.ObjectKey

	// Reference to a Secret containing a Root TLS Certificates used to verify the Relay Server Certificate.
	// The secret can also optionally specify a `tls.key` which will be used to generate the Agent Client Certificate.
	RootTlsSecret client.ObjectKey

	// Reference to a Secret containing a shared Token for authenticating to the Relay Server
	TokenSecret struct {
		Name      string
		Namespace string
		Key       string
	}
}

func (opts *Options) AddToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&opts.Server.Address, "relay-address", "", "address of the Relay gRPC Server (including port)")
	flags.BoolVar(&opts.Server.Insecure, "relay-insecure", false, "whether to connect over plaintext instead of HTTPS")
	flags.StringVar(&opts.Server.Authority, "relay-authority", "", "set the authority/host header to this value when dialing the Relay gRPC Server.")
	flags.BoolVar(&opts.Server.ReconnectOnNetworkFailures, "relay-backoff", true, "enable retry (with backoff) to reconnect to a disconnected server")
	flags.StringVar(&opts.Cluster, "cluster", "", "name of the cluster for which the agent will pull and resources. This should correspond to the name of the cluster as it is registered with the relay server.")
	flags.StringToStringVar(&opts.AgentLabels, "agent-labels", nil, "labels that will be applied to all resources output by this agent.")

	// relay identity
	flags.StringVar(&opts.ClientCertSecret.Name, "relay-client-cert-secret-name", "", "The name of the Secret containing the Client TLS Certificates used to identify the Relay Agent to the Server. If the secret does not exist, a Token and Root cert secret are required.")
	flags.StringVar(&opts.ClientCertSecret.Namespace, "relay-client-cert-secret-namespace", "", "The namespace of the Secret containing the Client TLS Certificates used to identify the Relay Agent to the Server. If the secret does not exist, a Token and Root cert secret are required.")

	flags.StringVar(&opts.RootTlsSecret.Name, "relay-root-cert-secret-name", "", "The name of the Secret containing a Root TLS Certificate used to verify the Relay Server Certificate. The secret can also optionally specify a `tls.key` which will be used to generate the Agent Client Certificate.")
	flags.StringVar(&opts.RootTlsSecret.Namespace, "relay-root-cert-secret-namespace", "", "The namespace of the Secret containing a Root TLS Certificate used to verify the Relay Server Certificate. The secret can also optionally specify a `tls.key` which will be used to generate the Agent Client Certificate.")

	flags.StringVar(&opts.TokenSecret.Name, "relay-identity-token-secret-name", "", "The name of the Secret containing the shared token used to authenticate to the Relay Server.")
	flags.StringVar(&opts.TokenSecret.Namespace, "relay-identity-token-secret-namespace", "", "The namespace of the Secret containing the shared token used to authenticate to the Relay Server.")
	flags.StringVar(&opts.TokenSecret.Key, "relay-identity-token-secret-key", "token", "The data key of the Secret containing the shared token used to authenticate to the Relay Server.")
}

// Copied from https://github.com/solo-io/skv2-enterprise/blob/b6945a04e5ff216469ccbc5feee0d201fc8266cc/relay/pkg/grpc/dialer.go#L146
// creates the initial TLS connection from agent to server for establishing identity
// by decorating grpcutils.DialOpts with TLS credentials derived from the rootCertSecretRef for verifying server identity
// exposed for usage in validation contexts
func DialIdentityServer(
	ctx context.Context,
	relayServerDialOpts grpcutils.DialOpts,
	secretClient v1.SecretClient,
	rootCertSecretRef client.ObjectKey,
) (*grpc.ClientConn, error) {
	rootCertSecret, err := secretClient.GetSecret(ctx, rootCertSecretRef)
	if err != nil {
		return nil, err
	}
	rootCert, ok := rootCertSecret.Data[corev1.ServiceAccountRootCAKey]
	if !ok {
		return nil, eris.Errorf("root cert secret missing tls.crt key")
	}

	// initialize identity client. this will use a short-lived connection to the grpc server which does not present a client cert
	initialServer := relayServerDialOpts

	tlsConfig, err := TLSConfig(rootCert, nil, nil, true)
	if err != nil {
		return nil, err
	}
	tlsConfig.ServerName = initialServer.Authority

	initialServer.ExtraOptions = append(
		initialServer.ExtraOptions,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)

	initialConnection, err := initialServer.Dial(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "dialing identity server failed")
	}
	return initialConnection, nil
}

// Copied from https://github.com/solo-io/skv2-enterprise/blob/4faa29210abf6591325ffe0bbdac09f1bda2befe/relay/pkg/identity/server/ca/secret.go#L50
// SecretToTLSConfig parses TLS config data from a k8s secret storing tls certs
func SecretToTLSConfig(secret *corev1.Secret, client bool) (*tls.Config, error) {
	tlsCrt := secret.Data[corev1.TLSCertKey]
	if len(tlsCrt) == 0 {
		return nil, eris.Errorf("missing ca cert")
	}
	tlsKey := secret.Data[corev1.TLSPrivateKeyKey]
	if len(tlsKey) == 0 {
		return nil, eris.Errorf("missing ca key")
	}
	root := secret.Data[corev1.ServiceAccountRootCAKey]
	if len(root) == 0 {
		root = tlsCrt
	}
	cfg, err := TLSConfig(root, tlsCrt, tlsKey, client)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Copied from https://github.com/solo-io/skv2-enterprise/blob/4faa29210abf6591325ffe0bbdac09f1bda2befe/relay/pkg/identity/utils/utils.go#L54-L85
func TLSConfig(root, certPem, keyPem []byte, client bool) (*tls.Config, error) {

	rootPem, _ := pem.Decode(root)

	rootcert, err := x509.ParseCertificate(rootPem.Bytes)
	if err != nil {
		return nil, err
	}
	roots := x509.NewCertPool()
	roots.AddCert(rootcert)

	var parsedCerts []tls.Certificate
	if len(certPem) > 0 {
		cert, err := ParseCert(certPem, keyPem)
		if err != nil {
			return nil, err
		}
		parsedCerts = []tls.Certificate{cert}
	}

	if client {
		return &tls.Config{
			RootCAs:      roots,
			Certificates: parsedCerts,
		}, nil
	} else {
		return &tls.Config{
			Certificates: parsedCerts,
			ClientCAs:    roots,
		}, nil
	}
}

// Copied from https://github.com/solo-io/skv2-enterprise/blob/4faa29210abf6591325ffe0bbdac09f1bda2befe/relay/pkg/identity/utils/utils.go#L46-L52
func ParseCert(certPem, keyPem []byte) (tls.Certificate, error) {
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return tls.Certificate{}, err
	}
	return cert, nil
}
