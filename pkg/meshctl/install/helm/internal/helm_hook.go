package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

/*
Copied with heavy modification from Helm, https://github.com/helm/helm/blob/041ce5a2c17a58be0fcd5f5e16fb3e7e95fea622/pkg/action/hooks.go#L30
*/

// execHook executes all of the hooks for the given hook event.
func ExecHook(
	rl *release.Release,
	hook release.HookEvent,
	timeout time.Duration,
	namespace string,
	out io.Writer,
	kubeRestClientGetter genericclioptions.RESTClientGetter,
) error {
	kubeClient := kube.New(kubeRestClientGetter)
	executingHooks := []*release.Hook{}

	restConfig, err := kubeRestClientGetter.ToRESTConfig()
	if err != nil {
		return err
	}
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	for _, h := range rl.Hooks {
		for _, e := range h.Events {
			if e == hook {
				executingHooks = append(executingHooks, h)
			}
		}
	}

	// hooke are pre-ordered by kind, so keep order stable
	sort.Stable(hookByWeight(executingHooks))

	for _, h := range executingHooks {
		resources, err := kubeClient.Build(bytes.NewBufferString(h.Manifest), true)
		if err != nil {
			return eris.Wrapf(err, "unable to build kubernetes object for %s hook %s", hook, h.Path)
		}

		// Create hook resources
		if _, err := kubeClient.Create(resources); err != nil {
			if errors.IsAlreadyExists(err) {
				return eris.Wrap(err, "Could not execute in-cluster test, delete the conflicting resources from the cluster and try again")
			}
			return eris.Wrapf(err, "Hook %s %s failed", hook, h.Path)
		}

		// Watch hook resources until they have completed
		err = kubeClient.WatchUntilReady(resources, timeout)
		if err != nil {
			return err
		}

		// fetch logs
		for _, e := range h.Events {
			if e == release.HookTest {

				req := client.CoreV1().Pods(namespace).GetLogs(h.Name, &v1.PodLogOptions{})
				logReader, err := req.Stream(context.Background())
				if err != nil {
					return eris.Wrapf(err, "unable to get pod logs for %s", h.Name)
				}

				fmt.Fprintf(out, "POD LOGS: %s\n", h.Name)
				_, err = io.Copy(out, logReader)
				fmt.Fprintln(out)
				if err != nil {
					return eris.Wrapf(err, "unable to write pod logs for %s", h.Name)
				}
			}
		}

		_, errs := kubeClient.Delete(resources)
		if len(errs) > 0 {
			return eris.New(joinErrors(errs))
		}
	}

	return nil
}

// hookByWeight is a sorter for hooks
type hookByWeight []*release.Hook

func (x hookByWeight) Len() int      { return len(x) }
func (x hookByWeight) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
func (x hookByWeight) Less(i, j int) bool {
	if x[i].Weight == x[j].Weight {
		return x[i].Name < x[j].Name
	}
	return x[i].Weight < x[j].Weight
}

func joinErrors(errs []error) string {
	es := make([]string, 0, len(errs))
	for _, e := range errs {
		es = append(es, e.Error())
	}
	return strings.Join(es, "; ")
}
