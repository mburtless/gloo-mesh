module github.com/solo-io/gloo-mesh

go 1.16

replace (
	// pinned to solo-io's fork of cue version 95a50cebaffb4bdba8c544601d8fb867990ad1ad
	cuelang.org/go => github.com/solo-io/cue v0.4.1-0.20210622213027-95a50cebaffb

	// github.com/Azure/go-autorest/autorest has different versions for the Go
	// modules than it does for releases on the repository. Note the correct
	// version when updating.
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	github.com/envoyproxy/go-control-plane => github.com/envoyproxy/go-control-plane v0.9.10-0.20210708144103-3a95f2df6351

	github.com/spf13/viper => github.com/istio/viper v1.3.3-0.20190515210538-2789fed3109c

	k8s.io/api => k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.2
	k8s.io/client-go => k8s.io/client-go v0.22.1
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.4.0
	k8s.io/kubectl => k8s.io/kubectl v0.22.2

)

require (
	contrib.go.opencensus.io/exporter/prometheus v0.4.0 // indirect
	cuelang.org/go v0.4.0
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d // indirect
	github.com/aws/aws-app-mesh-controller-for-k8s v1.1.1
	github.com/aws/aws-sdk-go v1.41.7
	github.com/briandowns/spinner v1.16.0
	github.com/cncf/udpa/go v0.0.0-20210930031921-04548b0d99d4 // indirect
	github.com/cncf/xds/go v0.0.0-20211011173535-cb28da3451f1 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/distribution/distribution/v3 v3.0.0-20210926092439-1563384b69df // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.7+incompatible // indirect
	github.com/envoyproxy/go-control-plane v0.9.10-0.20210708144103-3a95f2df6351
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/gertd/go-pluralize v0.1.1
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-test/deep v1.0.7
	github.com/gobuffalo/packr v1.30.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-github/v32 v32.0.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/iancoleman/strcase v0.1.3
	github.com/klauspost/compress v1.13.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/linkerd/linkerd2 v0.5.1-0.20200402173539-fee70c064bc0
	github.com/lucas-clemente/quic-go v0.24.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	github.com/openservicemesh/osm v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.32.1 // indirect
	github.com/pseudomuto/protoc-gen-doc v1.4.1
	github.com/pseudomuto/protokit v0.2.0
	github.com/rotisserie/eris v0.4.0
	github.com/servicemeshinterface/smi-sdk-go v0.4.1
	github.com/sirupsen/logrus v1.8.1
	github.com/solo-io/anyvendor v0.0.4
	github.com/solo-io/external-apis v0.1.10
	github.com/solo-io/go-list-licenses v0.1.3
	github.com/solo-io/go-utils v0.21.9
	github.com/solo-io/k8s-utils v0.0.11
	github.com/solo-io/protoc-gen-ext v0.0.16
	github.com/solo-io/skv2 v0.21.4
	github.com/solo-io/solo-apis v1.6.31
	github.com/solo-io/solo-kit v0.21.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	go.uber.org/atomic v1.9.0
	go.uber.org/zap v1.19.1
	golang.org/x/net v0.0.0-20211020060615-d418f374d309 // indirect
	golang.org/x/sys v0.0.0-20211020174200-9d6173849985 // indirect
	google.golang.org/api v0.59.0 // indirect
	google.golang.org/genproto v0.0.0-20211020151524-b7c3a969101a // indirect
	google.golang.org/grpc v1.41.0-dev
	google.golang.org/protobuf v1.27.1
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	helm.sh/helm/v3 v3.7.1
	istio.io/api v0.0.0-20211118145728-107306f7642e
	istio.io/client-go v1.12.0-beta.2.0.20211115195637-e27e6b79344f
	istio.io/gogo-genproto v0.0.0-20211115195057-0e34bdd2be67 // indirect
	istio.io/istio v0.0.0-20211013140753-9f6f03276054
	istio.io/pkg v0.0.0-20211115195056-e379f31ee62a
	istio.io/tools v0.0.0-20210420211536-9c0f48df3262
	k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/cli-runtime v0.22.2
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/klog/v2 v2.10.0 // indirect
	k8s.io/kube-openapi v0.0.0-20211020163157-7327e2aaee2b // indirect
	k8s.io/kubectl v0.22.2
	k8s.io/kubernetes v1.13.0
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b
	sigs.k8s.io/controller-runtime v0.10.2
	sigs.k8s.io/yaml v1.3.0 // indirect
)
