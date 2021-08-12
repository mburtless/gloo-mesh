package checks

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/rotisserie/eris"
	appsv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/gloo-mesh/pkg/certificates/common/secrets"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/checks/skv2enterprise"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/grpcutils"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type relayConnectivityCheck struct {
	preinstall bool
}

func NewRelayConnectivityCheck(
	preinstall bool,
) *relayConnectivityCheck {
	return &relayConnectivityCheck{
		preinstall: preinstall,
	}
}

func (r *relayConnectivityCheck) GetDescription() string {
	return "Gloo Mesh Agent Connectivity"
}

func (r *relayConnectivityCheck) Run(ctx context.Context, checkCtx CheckContext) *Result {
	result := &Result{}

	secretClient := checkCtx.Context().CoreClientset.Secrets()
	deploymentClient := checkCtx.Context().AppsClientset.Deployments()
	relayDialer := checkCtx.Context().RelayDialer

	// pre-install, verify root cert
	if r.preinstall {
		checkServerConnection(ctx, checkCtx.Context().AgentParams, relayDialer, result, secretClient)
	} else {

		// infer relay params from agent deployment
		agentParams, err := fetchAgentParamsFromDeployment(ctx, deploymentClient, checkCtx.Environment().Namespace)
		if err != nil {
			contextutils.LoggerFrom(ctx).DPanicf("could not read parameters from agent deployment: %v", err)
			return nil
		}

		checkServerConnection(ctx, agentParams, relayDialer, result, secretClient)
	}

	return result
}

// server connection check consists of:
//  * if it exists, verifying that the agent's client cert can be used to establish mTLS connection to the relay server
//  * else, verifying that agent's root cert can be used to establish initial TLS connection to the relay identity server
//  * verifying that relay server can be reached from agent's remote cluster
func checkServerConnection(
	ctx context.Context,
	agentParams *AgentParams,
	relayDialer RelayDialer,
	result *Result,
	secretClient v1.SecretClient,
) {
	relayServerDialOpts := grpcutils.DialOpts{
		Address:   agentParams.RelayServerAddress,
		Authority: agentParams.RelayAuthority,
		Insecure:  agentParams.Insecure,
	}

	// attempt insecure connection to relay server
	if relayServerDialOpts.Insecure {
		if err := relayDialer.DialServer(ctx, relayServerDialOpts); err != nil {
			if err == context.DeadlineExceeded {
				addTimeoutResult(result, relayServerDialOpts.Address)
			} else {
				result.AddError(eris.Wrap(err, "could not establish connection to relay server"))
			}
		}
		return
	}

	// attempt TLS or mTLS connection to relay server
	clientCertRef := agentParams.ClientCertSecretRef
	clientCertRefString := fmt.Sprintf("%s.%s", clientCertRef.Name, clientCertRef.Namespace)

	clientCertSecret, err := secretClient.GetSecret(ctx, clientCertRef)
	if err != nil && errors.IsNotFound(err) {
		// client cert does not exist, check the root cert instead
		checkDialingIdentity(ctx, relayDialer, result, secretClient, relayServerDialOpts, agentParams.RootCertSecretRef)
		return
	} else if err != nil {
		result.AddError(eris.Wrapf(err, "error reading the client certificate Secret: %s", clientCertRefString))
		return
	}

	// validate integrity of client cert
	tlsConfig := validateClientCert(result, clientCertSecret)
	if tlsConfig == nil {
		return
	}

	tlsConfig.ServerName = relayServerDialOpts.Authority
	relayServerDialOpts.ExtraOptions = append(relayServerDialOpts.ExtraOptions, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// attempt mTLS connection to relay server
	if err = relayDialer.DialServer(ctx, relayServerDialOpts); err != nil {
		if err == context.DeadlineExceeded {
			addTimeoutResult(result, relayServerDialOpts.Address)
		} else {
			result.AddError(eris.Wrap(err, "could not establish connection to relay server"))
			result.AddHint(
				fmt.Sprintf("validate the agent's client certificate (represented by the Secret %s)", clientCertRefString),
				"", // TODO: add documentation for manual creation of client TLS cert
			)
			result.AddHint(
				fmt.Sprintf("check that the relay server's address \"%s\" is correct", relayServerDialOpts.Address),
				"",
			)
		}
	}
}

// identity connection check consists of:
//   * verifying that relay server can be reached from agent's remote cluster
//   * verifying that the rootTlsSecret can be used to establish initial TLS connection to the relay server
func checkDialingIdentity(
	ctx context.Context,
	relayDialer RelayDialer,
	result *Result,
	secretClient v1.SecretClient,
	relayServerDialOpts grpcutils.DialOpts,
	rootCertSecretRef client.ObjectKey,
) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// attempt insecure connection to relay server
	if relayServerDialOpts.Insecure {
		if err := relayDialer.DialServer(ctx, relayServerDialOpts); err != nil {
			if err == context.DeadlineExceeded {
				addTimeoutResult(result, relayServerDialOpts.Address)
			} else {
				result.AddError(eris.Wrap(err, "could not establish connection to relay server"))
			}
		}
		return
	}

	// attempt tls connection to relay server
	if err := relayDialer.DialIdentity(ctx, relayServerDialOpts, secretClient, rootCertSecretRef); err != nil {
		if errors.IsNotFound(err) {
			result.AddError(eris.Errorf("the root certificate Secret does not exist: %s.%s", rootCertSecretRef.Name, rootCertSecretRef.Namespace))
		} else if err == context.DeadlineExceeded {
			addTimeoutResult(result, relayServerDialOpts.Address)
		} else {
			result.AddError(eris.Wrap(err, "could not establish connection to relay server"))
			result.AddHint(
				"check that the agent's root certificate Secret is valid",
				"https://docs.solo.io/gloo-mesh/latest/setup/installation/enterprise_installation/#manual-certificate-creation",
			)
			result.AddHint(
				fmt.Sprintf("check that the relay server's address (%s) is correct and can be reached from the agent's cluster", relayServerDialOpts.Address),
				"https://docs.solo.io/gloo-mesh/latest/setup/cluster_registration/enterprise_cluster_registration/#install-the-enterprise-agent",
			)
		}
	}
}

func addTimeoutResult(result *Result, relayAddress string) {
	result.AddError(eris.New("timed out attempting to connect to the relay server"))
	result.AddHint(
		fmt.Sprintf("check that the relay server's address (%s) is correct and can be reached from the agent's cluster", relayAddress),
		"https://docs.solo.io/gloo-mesh/latest/setup/cluster_registration/enterprise_cluster_registration/#install-the-enterprise-agent",
	)
}

// return tlsConfig if client cert is valid, otherwise return nil and mutate Result with error
func validateClientCert(
	result *Result,
	clientCertSecret *corev1.Secret,
) *tls.Config {
	tlsConfig, err := skv2enterprise.SecretToTLSConfig(clientCertSecret, true)
	if err != nil {
		invalidCertResult(result, err, sets.Key(clientCertSecret))
		return nil
	}

	// verify the integrity of the client cert
	certChainBytes := clientCertSecret.Data[corev1.TLSCertKey]
	decode, _ := pem.Decode(certChainBytes)
	certificates, err := x509.ParseCertificates(decode.Bytes)
	if err != nil {
		invalidCertResult(result, err, sets.Key(clientCertSecret))
		return nil
	}
	// the parsed certificates should only ever contain a single certificate
	cert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificates[0].Raw,
	})

	caData := secrets.CAData{
		RootCert:     clientCertSecret.Data[corev1.ServiceAccountRootCAKey],
		CertChain:    clientCertSecret.Data[corev1.TLSCertKey],
		CaCert:       cert,
		CaPrivateKey: clientCertSecret.Data[corev1.TLSPrivateKeyKey],
	}
	if err = caData.Verify(); err != nil {
		invalidCertResult(result, err, sets.Key(clientCertSecret))
		return nil
	}

	return tlsConfig
}

// mutate result with error indicating invalid cert
func invalidCertResult(result *Result, err error, clientCertSecretRefString string) {
	result.AddError(eris.Wrap(err, "the client certificate Secret is in an invalid format"))
	result.AddHint(
		fmt.Sprintf("validate the agent's client certificate (represented by the Secret %s)", clientCertSecretRefString),
		"", // TODO: add documentation for manual creation of client TLS cert
	)
}

// read the agent params from the deployment spec
func fetchAgentParamsFromDeployment(
	ctx context.Context,
	deploymentClient appsv1.DeploymentClient,
	installNamespace string,
) (*AgentParams, error) {
	// must match EnterpriseAgentChart.Data.Name, defined here https://github.com/solo-io/gloo-mesh-enterprise/blob/956d84b3cca461fe99433bbddc8bfd2bef0debfa/enterprise-networking/codegen/helm/chart.go#L121
	// we don't import it to avoid a dependency from this repo to GME
	enterpriseAgentName := "enterprise-agent"

	agentDeployment, err := deploymentClient.GetDeployment(ctx, client.ObjectKey{
		Name:      enterpriseAgentName,
		Namespace: installNamespace,
	})
	if err != nil {
		return nil, err
	}

	var agentContainer *corev1.Container
	for _, container := range agentDeployment.Spec.Template.Spec.Containers {
		if container.Name == enterpriseAgentName {
			agentContainer = &container
		}
	}
	if agentContainer == nil {
		return nil, eris.Errorf("no container found named %s", enterpriseAgentName)
	}

	// TODO: reuse the flag set defined for the GME agent
	// the agent's flag is a superset of the flags defined in skv2-enterprise for relay, ignore the extra flags
	flagSet := pflag.NewFlagSet("agent", pflag.ContinueOnError)
	flagSet.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{
		UnknownFlags: true,
	}
	agentOpts := skv2enterprise.Options{}
	agentOpts.AddToFlags(flagSet)
	if err = flagSet.Parse(agentContainer.Args); err != nil {
		return nil, err
	}

	return &AgentParams{
		RelayServerAddress:  agentOpts.Server.Address,
		RelayAuthority:      agentOpts.Server.Authority,
		Insecure:            agentOpts.Server.Insecure,
		RootCertSecretRef:   agentOpts.RootTlsSecret,
		ClientCertSecretRef: agentOpts.ClientCertSecret,
	}, nil
}

// wrap gRPC dial in interface for testing

//go:generate mockgen -source ./relay_connectivity.go -destination mocks/relay_connectivity.go

type RelayDialer interface {
	DialIdentity(
		ctx context.Context,
		relayServerDialOpts grpcutils.DialOpts,
		secretClient v1.SecretClient,
		rootCertSecretRef client.ObjectKey,
	) error

	DialServer(
		ctx context.Context,
		relayServerDialOpts grpcutils.DialOpts,
	) error
}

type relayDialer struct{}

func NewRelayDialer() RelayDialer {
	return &relayDialer{}
}

func (r *relayDialer) DialIdentity(
	ctx context.Context,
	relayServerDialOpts grpcutils.DialOpts,
	secretClient v1.SecretClient,
	rootCertSecretRef client.ObjectKey,
) error {
	_, err := skv2enterprise.DialIdentityServer(ctx, relayServerDialOpts, secretClient, rootCertSecretRef)
	return err
}

func (r *relayDialer) DialServer(
	ctx context.Context,
	relayServerDialOpts grpcutils.DialOpts,
) error {
	_, err := relayServerDialOpts.Dial(ctx)
	return err
}
