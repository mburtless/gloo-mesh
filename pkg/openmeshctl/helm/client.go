package helm

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/output"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
)

const (
	tempChartFilePermissions = 0644
)

// Client performs helm operations via a Kubernetes RESTFul client.
type Client struct {
	actionConfig *action.Configuration
	settings     *cli.EnvSettings
}

// NewClient returns a helm client configured to communicate with the Kubernetes cluster contained in the current
// context.
func NewClient(ctx runtime.Context) (*Client, error) {
	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(
		cli.New().RESTClientGetter(),
		ctx.Namespace(),
		os.Getenv("HELM_DRIVER"),
		func(s string, i ...interface{}) {
			fmt.Println(s)
		},
	); err != nil {
		return nil, err
	}

	return &Client{actionConfig, cli.New()}, nil
}

// Install creates a new release on the configured cluster with the provided chart.
func (c *Client) Install(ctx runtime.Context, spec ChartSpec, releaseName string, kubeContext string) error {
	client := action.NewInstall(c.actionConfig)

	client.CreateNamespace = true
	client.Namespace = spec.Namespace
	client.Version = spec.Version
	client.ReleaseName = releaseName
	client.IncludeCRDs = false

	out := os.Stderr

	if kubeContext == "" {
		var err error
		kubeContext, err = ctx.KubeContext()
		if err != nil {
			fmt.Printf("Error determining KubeContext - %v\n", err)
			return err
		}
	}

	settings := &cli.EnvSettings{
		KubeContext: kubeContext,
		KubeConfig:  ctx.KubeConfig(),
		Debug:       false,
	}

	valueOpts := spec.ValuesOptions

	fmt.Printf("Original chart version: %q\n", client.Version)
	if client.Version == "" && client.Devel {
		fmt.Printf("setting version to >0.0.0-0\n")
		client.Version = ">0.0.0-0"
	}

	chartRequested, err := downloadChart(spec.Name)
	if err != nil {
		fmt.Printf("Error downloading helm chart - %v\n", err)
		return err
	}

	cp := spec.Name

	p := getter.All(settings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		fmt.Printf("Error merging values - %v\n", err)
		return err
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              out,
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            settings.Debug,
				}
				if err := man.Update(); err != nil {
					fmt.Printf("Error updating dependencies - %v\n", err)
					return err
				}
				// Reload the chart with the updated Chart.lock file.
				if chartRequested, err = loader.Load(cp); err != nil {
					fmt.Printf("failed reloading chart after repo update - %v\n", err)
					return err
				}
			} else {
				fmt.Printf("Error checking dependencies - %v\n", err)
				return err
			}
		}
	}

	// Create context and prepare the handle of SIGTERM
	runCtx, cancel := context.WithCancel(ctx)

	// Handle SIGTERM
	cSignal := make(chan os.Signal)
	signal.Notify(cSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSignal
		fmt.Fprintf(out, "Release %s has been cancelled.\n", releaseName)
		cancel()
	}()

	rel, err := client.RunWithContext(runCtx, chartRequested, vals)

	if err != nil {
		fmt.Printf("error installing - %v\n", err)
	}

	var outfmt output.Format
	outfmt.Write(out, &statusPrinter{rel, settings.Debug, false})

	return nil
}

func (c *Client) getChart(chartPathOptions *action.ChartPathOptions, name string) (*chart.Chart, error) {
	chartPath, err := chartPathOptions.LocateChart(name, c.settings)
	if err != nil {
		return nil, err
	}

	return loader.Load(chartPath)
}

func (c *Client) updateDeps(ctx runtime.Context, chart *chart.Chart) error {
	deps := chart.Metadata.Dependencies
	if deps == nil {
		return nil
	}
	if err := action.CheckDependencies(chart, deps); err == nil {
		return nil
	} else {
		fmt.Printf("error checking dependencies: %s\n", err.Error())
	}
	man := &downloader.Manager{
		ChartPath:        chart.ChartPath(),
		Getters:          getter.All(c.settings),
		RepositoryConfig: c.settings.RepositoryConfig,
		RepositoryCache:  c.settings.RepositoryCache,
	}
	if err := man.Update(); err != nil {
		return eris.Wrap(err, "unable to update dependencies")
	}

	return nil
}

// Uninstall removes a release from the configured cluster.
func (c *Client) Uninstall(ctx runtime.Context, releaseName string) error {
	uninstaller := action.NewUninstall(c.actionConfig)
	// uninstaller.DryRun = ctx.DryRun
	if _, err := uninstaller.Run(releaseName); err != nil {
		return eris.Wrap(err, "unable to uninstall chart")
	}

	return nil
}

type statusPrinter struct {
	release         *release.Release
	debug           bool
	showDescription bool
}

func (s statusPrinter) WriteJSON(out io.Writer) error {
	return output.EncodeJSON(out, s.release)
}

func (s statusPrinter) WriteYAML(out io.Writer) error {
	return output.EncodeYAML(out, s.release)
}

func (s statusPrinter) WriteTable(out io.Writer) error {
	if s.release == nil {
		return nil
	}
	fmt.Fprintf(out, "NAME: %s\n", s.release.Name)
	if !s.release.Info.LastDeployed.IsZero() {
		fmt.Fprintf(out, "LAST DEPLOYED: %s\n", s.release.Info.LastDeployed.Format(time.ANSIC))
	}
	fmt.Fprintf(out, "NAMESPACE: %s\n", s.release.Namespace)
	fmt.Fprintf(out, "STATUS: %s\n", s.release.Info.Status.String())
	fmt.Fprintf(out, "REVISION: %d\n", s.release.Version)
	if s.showDescription {
		fmt.Fprintf(out, "DESCRIPTION: %s\n", s.release.Info.Description)
	}

	executions := executionsByHookEvent(s.release)
	if tests, ok := executions[release.HookTest]; !ok || len(tests) == 0 {
		fmt.Fprintln(out, "TEST SUITE: None")
	} else {
		for _, h := range tests {
			// Don't print anything if hook has not been initiated
			if h.LastRun.StartedAt.IsZero() {
				continue
			}
			fmt.Fprintf(out, "TEST SUITE:     %s\n%s\n%s\n%s\n",
				h.Name,
				fmt.Sprintf("Last Started:   %s", h.LastRun.StartedAt.Format(time.ANSIC)),
				fmt.Sprintf("Last Completed: %s", h.LastRun.CompletedAt.Format(time.ANSIC)),
				fmt.Sprintf("Phase:          %s", h.LastRun.Phase),
			)
		}
	}

	if s.debug {
		fmt.Fprintln(out, "USER-SUPPLIED VALUES:")
		err := output.EncodeYAML(out, s.release.Config)
		if err != nil {
			return err
		}
		// Print an extra newline
		fmt.Fprintln(out)

		cfg, err := chartutil.CoalesceValues(s.release.Chart, s.release.Config)
		if err != nil {
			return err
		}

		fmt.Fprintln(out, "COMPUTED VALUES:")
		err = output.EncodeYAML(out, cfg.AsMap())
		if err != nil {
			return err
		}
		// Print an extra newline
		fmt.Fprintln(out)
	}

	if strings.EqualFold(s.release.Info.Description, "Dry run complete") || s.debug {
		fmt.Fprintln(out, "HOOKS:")
		for _, h := range s.release.Hooks {
			fmt.Fprintf(out, "---\n# Source: %s\n%s\n", h.Path, h.Manifest)
		}
		fmt.Fprintf(out, "MANIFEST:\n%s\n", s.release.Manifest)
	}

	if len(s.release.Info.Notes) > 0 {
		fmt.Fprintf(out, "NOTES:\n%s\n", strings.TrimSpace(s.release.Info.Notes))
	}
	return nil
}

func executionsByHookEvent(rel *release.Release) map[release.HookEvent][]*release.Hook {
	result := make(map[release.HookEvent][]*release.Hook)
	for _, h := range rel.Hooks {
		for _, e := range h.Events {
			executions, ok := result[e]
			if !ok {
				executions = []*release.Hook{}
			}
			result[e] = append(executions, h)
		}
	}
	return result
}

func downloadChart(chartArchiveUri string) (*chart.Chart, error) {
	charFilePath := ""
	if fi, err := os.Stat(chartArchiveUri); err == nil && fi.IsDir() {
		charFilePath = chartArchiveUri
	} else {

		// 1. Get a reader to the chart file (remote URL or local file path)
		chartFileReader, err := getResource(chartArchiveUri)
		if err != nil {
			return nil, err
		}
		defer func() { _ = chartFileReader.Close() }()

		// 2. Write chart to a temporary file
		chartBytes, err := ioutil.ReadAll(chartFileReader)
		if err != nil {
			return nil, err
		}

		chartFile, err := ioutil.TempFile("", "temp-helm-chart")
		if err != nil {
			return nil, err
		}
		charFilePath = chartFile.Name()
		defer func() { _ = os.RemoveAll(charFilePath) }()

		if err := ioutil.WriteFile(charFilePath, chartBytes, tempChartFilePermissions); err != nil {
			return nil, err
		}
	}

	// 3. Load the chart file
	chartObj, err := loader.Load(charFilePath)
	if err != nil {
		return nil, err
	}

	return chartObj, nil
}

// Get the resource identified by the given URI.
// The URI can either be an http(s) address or a relative/absolute file path.
func getResource(uri string) (io.ReadCloser, error) {
	var file io.ReadCloser
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		resp, err := http.Get(uri)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, eris.Errorf("http GET returned status %d for resource %s", resp.StatusCode, uri)
		}

		file = resp.Body
	} else {
		path, err := filepath.Abs(uri)
		if err != nil {
			return nil, eris.Wrapf(err, "getting absolute path for %v", uri)
		}

		f, err := os.Open(path)
		if err != nil {
			return nil, eris.Wrapf(err, "opening file %v", path)
		}
		file = f
	}

	// Write the body to file
	return file, nil
}
