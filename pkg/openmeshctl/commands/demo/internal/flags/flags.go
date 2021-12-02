package flags

import (
	"github.com/spf13/pflag"
)

type Options struct {
	Version       string
	Chart         string
	SkipGMInstall bool
}

func (o *Options) AddToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.Version, "version", "",
		"Gloo Mesh version to install. defaults to meshctl version")
	flags.StringVar(&o.Chart, "chart", "",
		"Gloo Mesh helm chart to install on the management plane")
	flags.BoolVar(&o.SkipGMInstall, "skip-gm-install", false,
		"If set to true, the local kind clusters, Istio installation, and bookinfo applications are all installed - but Gloo Mesh is NOT installed. Useful for simluating an example environment to use for trying out manual installation of Gloo Mesh.",
	)
}
