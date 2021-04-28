module k8c.io/kubeone

go 1.16

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/MakeNowJust/heredoc/v2 v2.0.1
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/Masterminds/sprig/v3 v3.2.0
	github.com/aws/aws-sdk-go v1.36.2
	github.com/docker/distribution v2.7.1+incompatible
	github.com/dominodatalab/os-release v0.0.0-20190522011736-bcdb4a3e3c2f
	github.com/google/go-cmp v0.5.4
	github.com/imdario/mergo v0.3.11
	github.com/koron-go/prefixw v0.0.0-20181013140428-271b207a7572
	github.com/kubermatic/machine-controller v1.28.0
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	go.etcd.io/etcd/v3 v3.3.0-rc.0.0.20200728214110-6c81b20ec8de
	golang.org/x/crypto v0.0.0-20201217014255-9d1352758620
	golang.org/x/term v0.0.0-20201117132131-f5c789dd3221
	google.golang.org/grpc v1.33.2
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.19.4
	k8s.io/apiextensions-apiserver v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/apiserver v0.19.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/cluster-bootstrap v0.19.4
	k8s.io/code-generator v0.19.4
	k8s.io/kube-aggregator v0.19.4
	k8s.io/kubelet v0.19.4
	sigs.k8s.io/controller-runtime v0.7.2
	sigs.k8s.io/yaml v1.2.0
)

replace (
	google.golang.org/grpc => google.golang.org/grpc v1.27.1
	k8s.io/client-go => k8s.io/client-go v0.19.4
)
