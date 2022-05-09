package e2e

import (
	"fmt"
	"os"

	"k8c.io/kubeone/test/e2e/testutil"
)

var (
	defaultTFEnvironment = []string{
		"TF_IN_AUTOMATION=true",
		"TF_CLI_ARGS=-no-color",
	}
)

type terraformBin struct {
	name string
	path string
	vars []string
}

func (tf *terraformBin) init(name string) error {
	tf.name = name

	return tf.run("init")
}

func (tf *terraformBin) apply(additionaArgs ...string) error {
	args := []string{"apply", "-auto-approve"}

	for _, arg := range append(tf.vars, fmt.Sprintf("cluster_name=%s", tf.name)) {
		args = append(args, "-var", arg)
	}

	return tf.run(append(args, additionaArgs...)...)
}

func (tf *terraformBin) destroy() error {
	return tf.run("destroy", "-auto-approve")
}

func (tf *terraformBin) run(args ...string) error {
	return tf.build(args...).Run()
}

func (tf *terraformBin) build(args ...string) *testutil.Exec {
	return testutil.NewExec("terraform",
		testutil.WithArgs(args...),
		testutil.WithEnv(append(os.Environ(), defaultTFEnvironment...)),
		testutil.InDir(tf.path),
		testutil.WithDryRun(),
		testutil.StdoutDebug,
	)
}
