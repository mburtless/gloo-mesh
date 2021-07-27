package dashboard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/browser"
	corev1client "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	pkgclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var ConsoleNotFoundError = errors.New("Console image not found. Your Gloo Mesh enterprise install may be a bad state.")

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Port forwards the Gloo Mesh Enterprise UI and opens it in a browser if available",
		RunE: func(cmd *cobra.Command, args []string) error {
			return forwardDashboard(ctx, opts.kubeconfig, opts.kubecontext, opts.namespace, opts.port)
		},
	}
	opts.addToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	return cmd
}

type options struct {
	kubeconfig  string
	kubecontext string
	namespace   string
	port        uint32
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.StringVar(&o.namespace, "namespace", "gloo-mesh", "The namespace that the Gloo Mesh UI is deployed in")
	flags.Uint32VarP(&o.port, "port", "p", 8090, "The local port to forward to the dashboard")
}

func forwardDashboard(ctx context.Context, kubeconfigPath, kubectx, namespace string, localPort uint32) error {
	staticPort, err := getStaticPort(ctx, kubeconfigPath, kubectx, namespace)
	if err != nil {
		return err
	}
	portFwdCmd, err := forwardPort(kubeconfigPath, kubectx, namespace, fmt.Sprint(localPort), staticPort)
	if err != nil {
		return err
	}
	defer portFwdCmd.Wait()
	if err := browser.OpenURL(fmt.Sprintf("http://localhost:%d", localPort)); err != nil {
		return err
	}

	return nil
}

func getStaticPort(ctx context.Context, kubeconfigPath, kubectx, namespace string) (string, error) {
	client, err := utils.BuildClient(kubeconfigPath, kubectx)
	if err != nil {
		return "", err
	}
	svcClient := corev1client.NewServiceClient(client)
	svc, err := svcClient.GetService(ctx, pkgclient.ObjectKey{Name: "dashboard", Namespace: namespace})
	if apierrors.IsNotFound(err) {
		fmt.Printf("No Gloo Mesh dashboard found as part of the installation in namespace %s. "+
			"The full dashboard is part of Gloo Mesh enterprise by default. "+
			"Check that your kubeconfig is pointing at the Gloo Mesh management cluster. ", namespace)
	} else if err != nil {
		return "", err
	}

	for _, port := range svc.Spec.Ports {
		if port.Name == "console" {
			return port.TargetPort.String(), nil
		}
	}

	return "", ConsoleNotFoundError
}

func forwardPort(kubeconfigPath, kubectx, namespace, localPort, kubePort string) (*exec.Cmd, error) {
	cmdArgs := []string{
		"port-forward",
		"-n",
		namespace,
		"deployment/dashboard",
		localPort + ":" + kubePort,
	}
	if kubectx != "" {
		cmdArgs = append(cmdArgs, "--context", kubectx)
	}
	if kubeconfigPath != "" {
		cmdArgs = append(cmdArgs, "--kubeconfig", kubeconfigPath)
	}
	cmd := exec.Command("kubectl", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	if err := waitForDashboard(cmd, localPort); err != nil {
		return nil, err
	}

	return cmd, nil
}

func waitForDashboard(portFwdCmd *exec.Cmd, localPort string) error {
	ticker, timer := time.NewTicker(250*time.Millisecond), time.NewTimer(30*time.Second)
	errs := &multierror.Error{}
	for {
		err := func() error {
			res, err := http.Get("http://localhost:" + localPort)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("invalid status code: %d %s", res.StatusCode, res.Status)
			}
			io.Copy(ioutil.Discard, res.Body)
			return nil
		}()
		if err == nil {
			return nil
		}

		errs = multierror.Append(errs, err)

		select {
		case <-timer.C:
			if portFwdCmd.Process != nil {
				portFwdCmd.Process.Kill()
				portFwdCmd.Process.Release()
			}

			return fmt.Errorf("timed out waiting for dashboard port forward to be ready: %s", errs.Error())
		case <-ticker.C:
		}
	}
}
