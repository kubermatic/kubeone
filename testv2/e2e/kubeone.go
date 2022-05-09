package e2e

import (
	"os"

	"k8c.io/kubeone/test/e2e/testutil"
)

type kubeoneBin struct {
	bin             string
	dir             string
	tfjsonPath      string
	manifestPath    string
	credentialsPath string
}

func (k1 *kubeoneBin) globalFlags() []string {
	args := []string{"--verbose", "--tfjson", k1.tfjsonPath}

	if k1.manifestPath != "" {
		args = append(args, "--manifest", k1.manifestPath)
	}

	if k1.credentialsPath != "" {
		args = append(args, "--credentials", k1.credentialsPath)
	}

	return args
}

func (k1 *kubeoneBin) Apply() error {
	return k1.run("apply", "--auto-approve")
}

func (k1 *kubeoneBin) Kubeconfig() error {
	return k1.run("kubeconfig")
}

func (k1 *kubeoneBin) Reset() error {
	return k1.run("reset", "--auto-approve", "--destroy-workers")
}

func (k1 *kubeoneBin) run(args ...string) error {
	return k1.build(append(k1.globalFlags(), args...)...).Run()
}

func (k1 *kubeoneBin) build(args ...string) *testutil.Exec {
	bin := "kubeone"
	if k1.bin != "" {
		bin = k1.bin
	}

	return testutil.NewExec(bin,
		testutil.WithArgs(args...),
		testutil.WithEnv(os.Environ()),
		testutil.InDir(k1.dir),
		testutil.WithDryRun(),
		testutil.StdoutDebug,
	)
}
