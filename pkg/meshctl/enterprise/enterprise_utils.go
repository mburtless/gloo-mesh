package enterprise

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	. "github.com/logrusorgru/aurora/v3"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultRelayAuthority = "enterprise-networking.gloo-mesh"
	DefaultRootCAName     = "relay-root-tls-secret"
	DefaultClientCertName = "relay-client-tls-secret"
	DefaultTokenName      = "relay-identity-token-secret"
	DefaultTokenSecretKey = "token"
)

type RegistrationOptions struct {
	registration.Options
	AgentChartPathOverride string
	AgentChartValuesPath   string

	RelayServerAddress  string
	RelayServerInsecure bool

	RootCASecretName      string
	RootCASecretNamespace string

	ClientCertSecretName      string
	ClientCertSecretNamespace string

	TokenSecretName      string
	TokenSecretNamespace string
	TokenSecretKey       string

	ReleaseName string

	SkipChecks bool
}

// construct the helm installer from the specified options
func (o *RegistrationOptions) GetInstaller(ctx context.Context, out io.Writer) (*helm.Installer, error) {
	chartPath, err := o.GetChartPath(ctx, o.AgentChartPathOverride, gloomesh.EnterpriseAgentChartUriTemplate)
	if err != nil {
		return nil, err
	}

	releaseName := gloomesh.EnterpriseAgentReleaseName
	if o.ReleaseName != "" {
		releaseName = o.ReleaseName
	}

	values := map[string]string{
		"relay.serverAddress": o.RelayServerAddress,
		"global.insecure":     strconv.FormatBool(o.RelayServerInsecure),
		"relay.cluster":       o.ClusterName,
	}

	if !o.RelayServerInsecure {
		values["relay.rootTlsSecret.name"] = o.RootCASecretName
		values["relay.rootTlsSecret.namespace"] = o.RootCASecretNamespace

		// relay needs a client cert name provided, even if it doesn't exist, so it can upsert the Secret if needed
		if o.ClientCertSecretName == "" {
			o.ClientCertSecretName = DefaultClientCertName
		}
		if o.ClientCertSecretNamespace == "" {
			o.ClientCertSecretNamespace = o.RemoteNamespace
		}
		values["relay.clientCertSecret.name"] = o.ClientCertSecretName
		values["relay.clientCertSecret.namespace"] = o.ClientCertSecretNamespace

		// only copy token secret if we have one
		if o.TokenSecretName != "" {
			values["relay.tokenSecret.name"] = o.TokenSecretName
			values["relay.tokenSecret.namespace"] = o.TokenSecretNamespace
			values["relay.tokenSecret.key"] = o.TokenSecretKey
		}
	}

	return &helm.Installer{
		KubeConfig:  o.KubeConfigPath,
		KubeContext: o.RemoteContext,
		ChartUri:    chartPath,
		Namespace:   o.RemoteNamespace,
		ReleaseName: releaseName,
		ValuesFile:  o.AgentChartValuesPath,
		Verbose:     o.Verbose,
		Values:      values,
		Output:      out,
	}, nil
}

func ensureCerts(ctx context.Context, opts *RegistrationOptions) (bool, error) {
	if opts.RootCASecretName != "" && (opts.ClientCertSecretName != "" || opts.TokenSecretName != "") {
		// we have all the data we need: root ca and either a client cert or a token.
		// nothing to be done here
		return false, nil
	}

	createdBootstrapToken := false

	mgmtKubeConfigPath := opts.KubeConfigPath
	// override if provided
	if opts.MgmtKubeConfigPath != "" {
		mgmtKubeConfigPath = opts.MgmtKubeConfigPath
	}
	mgmtKubeClient, err := utils.BuildClient(mgmtKubeConfigPath, opts.MgmtContext)
	if err != nil {
		return createdBootstrapToken, err
	}
	remoteKubeClient, err := utils.BuildClient(opts.KubeConfigPath, opts.RemoteContext)
	if err != nil {
		return createdBootstrapToken, err
	}
	mgmtKubeSecretClient := v1.NewSecretClient(mgmtKubeClient)
	remoteKubeSecretClient := v1.NewSecretClient(remoteKubeClient)

	if opts.RootCASecretName == "" {
		opts.RootCASecretName = DefaultRootCAName
		if opts.RootCASecretNamespace == "" {
			opts.RootCASecretNamespace = opts.RemoteNamespace
		}
		mgmtRootCaNameNamespace := client.ObjectKey{
			Name:      opts.RootCASecretName,
			Namespace: opts.MgmtNamespace,
		}

		if err = utils.EnsureNamespace(ctx, remoteKubeClient, opts.RemoteNamespace); err != nil {
			return createdBootstrapToken, eris.Wrapf(err, "creating namespace")
		}
		// no root cert, try copy it over
		logrus.Info("📃 Copying root CA ", Bold(fmt.Sprintf("%s.%s", mgmtRootCaNameNamespace.Name, mgmtRootCaNameNamespace.Namespace)), " to remote cluster from management cluster")

		s, err := mgmtKubeSecretClient.GetSecret(ctx, mgmtRootCaNameNamespace)
		if err != nil {
			return createdBootstrapToken, err
		}
		copiedSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      opts.RootCASecretName,
				Namespace: opts.RootCASecretNamespace,
			},
			Data: map[string][]byte{
				"ca.crt": s.Data["ca.crt"],
			},
		}
		// Write it to the remote cluster. Note that we use create to make sure we don't overwrite
		// anything that already exist.
		err = remoteKubeSecretClient.CreateSecret(ctx, copiedSecret)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return createdBootstrapToken, err
		}
	}

	if opts.ClientCertSecretName != "" {
		// if we now have a client cert, we have everything we need
		return createdBootstrapToken, nil
	}

	if opts.TokenSecretName == "" {
		// no token, copy it from mgmt cluster:
		opts.TokenSecretName = DefaultTokenName
		if opts.TokenSecretNamespace == "" {
			opts.TokenSecretNamespace = opts.RemoteNamespace
		}
		if opts.TokenSecretKey == "" {
			opts.TokenSecretKey = DefaultTokenSecretKey
		}
		mgmtTokenNameNamespace := client.ObjectKey{
			Name:      opts.TokenSecretName,
			Namespace: opts.MgmtNamespace,
		}
		if err = utils.EnsureNamespace(ctx, remoteKubeClient, opts.RemoteNamespace); err != nil {
			return createdBootstrapToken, eris.Wrapf(err, "creating namespace")
		}
		logrus.Info("📃 Copying bootstrap token ", Bold(fmt.Sprintf("%s.%s", opts.TokenSecretName, opts.TokenSecretNamespace)), " to remote cluster from management cluster")
		// no root cert, try copy it over
		s, err := mgmtKubeSecretClient.GetSecret(ctx, mgmtTokenNameNamespace)
		if err != nil {
			return createdBootstrapToken, err
		}
		copiedSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      opts.TokenSecretName,
				Namespace: opts.TokenSecretNamespace,
			},
			Data: s.Data,
		}
		// write it to the remote cluster
		err = remoteKubeSecretClient.CreateSecret(ctx, copiedSecret)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			// A write error occurred.
			return createdBootstrapToken, err
		} else if err != nil && apierrors.IsAlreadyExists(err) {
			// The user either provisioned their own token secret, or
			// we're on the management cluster and using the server's
			// token secret.
			createdBootstrapToken = false
		} else {
			// Successfully created the bootstrap token on a remote cluster.
			createdBootstrapToken = true
		}
	}
	return createdBootstrapToken, nil
}

// registers a new cluster with the provided automation:
//   * creates relay certs if they don't exist
//   * runs agent pre install checks
func RegisterCluster(ctx context.Context, opts RegistrationOptions) error {
	bootstrapTokenCreated := false
	if !opts.RelayServerInsecure {
		var err error
		// ensure existence of relay certs, create if needed
		bootstrapTokenCreated, err = ensureCerts(ctx, &opts)
		if err != nil {
			return err
		}
	}

	installer, err := opts.GetInstaller(ctx, os.Stdout)

	// wait for all enterprise-agent resources to be ready
	installer.Wait = true

	if err != nil {
		return eris.Wrap(err, "error building installer")
	}

	logrus.Info("💻 Installing relay agent in the remote cluster")

	if err := installer.InstallChart(ctx); err != nil {
		return err
	}

	kubeConfigPath := opts.MgmtKubeConfigPath
	if kubeConfigPath == "" {
		kubeConfigPath = opts.KubeConfigPath
	}

	mgmtKubeClient, err := utils.BuildClient(kubeConfigPath, opts.MgmtContext)
	if err != nil {
		return err
	}
	mgmtClusterClient := v1alpha1.NewKubernetesClusterClient(mgmtKubeClient)

	logrus.Info("📃 Creating ", Bold(opts.ClusterName+" KubernetesCluster CRD"), " in management cluster")

	err = mgmtClusterClient.CreateKubernetesCluster(ctx, &v1alpha1.KubernetesCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.ClusterName,
			Namespace: opts.MgmtNamespace,
		},
		Spec: v1alpha1.KubernetesClusterSpec{
			ClusterDomain: opts.ClusterDomain,
		},
	})
	if err != nil {
		return err
	}

	if !opts.RelayServerInsecure {
		logrus.Info("⌚ Waiting for relay agent to have a client certificate")

		remoteKubeClient, err := utils.BuildClient(opts.KubeConfigPath, opts.RemoteContext)
		if err != nil {
			return err
		}
		remoteKubeSecretClient := v1.NewSecretClient(remoteKubeClient)

		err = waitForClientCert(ctx, remoteKubeSecretClient, opts)
		if err != nil {
			return err
		}
		if bootstrapTokenCreated {
			// Delete the bootstrap token from the registered cluster
			// if it was created by this command invocation.
			logrus.Info("🗑 Removing bootstrap token")
			key := client.ObjectKey{
				Name:      opts.TokenSecretName,
				Namespace: opts.TokenSecretNamespace,
			}
			err = remoteKubeSecretClient.DeleteSecret(ctx, key)
			if err != nil {
				return err
			}
		}
	}

	logrus.Info("✅  Done registering cluster!")

	// agent post install checks
	if !opts.SkipChecks {
		logrus.Info("🔎 Performing agent post-install checks...")
		if err := installer.ExecuteHelmTest(); err != nil {
			return eris.Wrap(err, "agent post-install check failed")
		}
		logrus.Info("✅  Agent post-install checks succeeded!")
	}

	return nil
}

func waitForClientCert(ctx context.Context, remoteKubeSecretClient v1.SecretClient, opts RegistrationOptions) error {

	clientCert := client.ObjectKey{
		Name:      opts.ClientCertSecretName,
		Namespace: opts.ClientCertSecretNamespace,
	}

	timeout := time.After(2 * time.Minute)

	for {
		_, err := remoteKubeSecretClient.GetSecret(ctx, clientCert)
		if !apierrors.IsNotFound(err) {
			return err
		}
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return eris.Errorf("timed out waiting for client cert")
		case <-time.After(5 * time.Second):
			logrus.Info("\t Checking...")
		}
	}
	logrus.Info("📃 Client certificate found in remote cluster")
	return nil
}

func DeregisterCluster(ctx context.Context, opts RegistrationOptions) error {

	logrus.Infof("deleting KubernetesCluster CR %s.%s from management cluster...", opts.ClusterName, opts.MgmtNamespace)

	kubeClient, err := utils.BuildClient(opts.KubeConfigPath, opts.MgmtContext)
	if err != nil {
		return err
	}
	clusterKey := client.ObjectKey{Name: opts.ClusterName, Namespace: opts.MgmtNamespace}
	if err = v1alpha1.NewKubernetesClusterClient(kubeClient).DeleteKubernetesCluster(ctx, clusterKey); err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	logrus.Info("uninstalling enterprise agent from remote cluster...")
	releaseName := gloomesh.EnterpriseAgentReleaseName
	if opts.ReleaseName != "" {
		releaseName = opts.ReleaseName
	}
	return (helm.Uninstaller{
		KubeConfig:  opts.KubeConfigPath,
		KubeContext: opts.RemoteContext,
		Namespace:   opts.RemoteNamespace,
		ReleaseName: releaseName,
		Verbose:     opts.Verbose,
	}).UninstallChart(ctx)
}
