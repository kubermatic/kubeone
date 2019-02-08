package util

import (
	"github.com/sirupsen/logrus"

	"github.com/kubermatic/kubeone/pkg/config"
	"github.com/kubermatic/kubeone/pkg/ssh"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Context hold together currently test flags and parsed info, along with
// utilities like logger
type Context struct {
	Cluster               *config.Cluster
	Logger                logrus.FieldLogger
	Connector             *ssh.Connector
	Configuration         *Configuration
	Runner                *Runner
	WorkDir               string
	JoinCommand           string
	JoinToken             string
	Clientset             *kubernetes.Clientset
	APIExtensionClientset *apiextensionsclientset.Clientset
	RESTConfig            *rest.Config
	Verbose               bool
	BackupFile            string
	DestroyWorkers        bool
}

// Clone returns a shallow copy of the context.
func (c *Context) Clone() *Context {
	newCtx := *c
	return &newCtx
}
