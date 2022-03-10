/*
Copyright 2019 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"

	"k8c.io/kubeone/pkg/templates/images"

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
func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display KubeOne version",
		Long: heredoc.Doc(`
			Prints the exact version number, as embedded by the build system.
		`),
		SilenceErrors: true,
		Args:          cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			ownver := k8sversion.Info{
				GitVersion: version,
				GitCommit:  commit,
				BuildDate:  date,
				Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
				Compiler:   runtime.Compiler,
				GoVersion:  runtime.Version(),
			}

			imgResolver := images.NewResolver()
			mcTag := imgResolver.Tag(images.MachineController)

			mcver := k8sversion.Info{
				GitVersion: mcTag,
				Platform:   "linux/amd64",
			}

			ownsver, err := semver.NewVersion(version)
			if err == nil {
				ownver.Major = strconv.Itoa(int(ownsver.Major()))
				ownver.Minor = strconv.Itoa(int(ownsver.Minor()))
			}

			mcsver, err := semver.NewVersion(mcTag)
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
