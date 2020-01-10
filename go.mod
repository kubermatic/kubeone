module github.com/kubermatic/kubeone

go 1.13

require (
	github.com/Masterminds/semver v1.4.2
	github.com/aws/aws-sdk-go v1.20.15
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/imdario/mergo v0.3.7
	github.com/koron-go/prefixw v0.0.0-20181013140428-271b207a7572
	github.com/kr/fs v0.1.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de
	github.com/pkg/errors v0.8.1
	github.com/pkg/sftp v1.10.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.15.7
	k8s.io/apiextensions-apiserver v0.15.7
	k8s.io/apimachinery v0.15.7
	k8s.io/client-go v0.15.7
	k8s.io/cluster-bootstrap v0.15.7
	k8s.io/code-generator v0.15.7
	k8s.io/gengo v0.0.0-20191010091904-7fa3014cb28f // indirect
	k8s.io/kube-aggregator v0.15.7
	sigs.k8s.io/cluster-api v0.0.0-20190603191137-2ec456177c0e
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/yaml v1.1.0
)
