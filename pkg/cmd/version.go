package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kubermatic/kubeone/pkg/templates/machinecontroller"

	k8sversion "k8s.io/apimachinery/pkg/version"
)

var (
	commit  = "none"
	date    = "unknown"
	version = "dev"
)

type kubeoneVersions struct {
	Kubeone           k8sversion.Info `json:"kubeone"`
	MachineController k8sversion.Info `json:"machine_controller"`
}

// versionCmd setups version command
func versionCmd(_ *pflag.FlagSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display KubeOne version",
		Long:  `Prints the exact version number, as embedded by the build system.`,
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			ownver := k8sversion.Info{
				GitVersion: version,
				GitCommit:  commit,
				BuildDate:  date,
				Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
				Compiler:   runtime.Compiler,
				GoVersion:  runtime.Version(),
			}

			mcver := k8sversion.Info{
				GitVersion: machinecontroller.MachineControllerTag,
				Platform:   "linux/amd64",
			}

			ownsver, err := semver.NewVersion(version)
			if err == nil {
				ownver.Major = strconv.Itoa(int(ownsver.Major()))
				ownver.Minor = strconv.Itoa(int(ownsver.Minor()))
			}

			mcsver, err := semver.NewVersion(machinecontroller.MachineControllerTag)
			if err == nil {
				mcver.Major = strconv.Itoa(int(mcsver.Major()))
				mcver.Minor = strconv.Itoa(int(mcsver.Minor()))
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")

			return enc.Encode(kubeoneVersions{
				Kubeone:           ownver,
				MachineController: mcver,
			})
		},
	}

	return cmd
}
