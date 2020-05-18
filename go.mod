module github.com/kubermatic/kubeone

go 1.13

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.4.2
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/aws/aws-sdk-go v1.20.15
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/imdario/mergo v0.3.8
	github.com/koron-go/prefixw v0.0.0-20181013140428-271b207a7572
	github.com/kubermatic/machine-controller v1.11.1
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.11.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738 // 3cf2f69b5738 is the SHA for git tag v3.4.3
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/crypto v0.0.0-20200214034016-1d94cc7ab1c6
	google.golang.org/grpc v1.27.1
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery v0.16.4
	k8s.io/client-go v0.16.4
	k8s.io/cluster-bootstrap v0.16.4
	k8s.io/code-generator v0.16.4
	k8s.io/kube-aggregator v0.16.4
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/yaml v1.2.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0 // fixed for etcd to compile
