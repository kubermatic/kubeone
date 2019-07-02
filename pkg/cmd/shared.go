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
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/cluster"
)

const (
	globalTerraformFlagName = "tfjson"
	globalVerboseFlagName   = "verbose"
	globalDebugFlagName     = "debug"
)

// globalOptions are global globalOptions same for all commands
type globalOptions struct {
	TerraformState string
	Verbose        bool
	Debug          bool
}

func persistentGlobalOptions(fs *pflag.FlagSet) (*globalOptions, error) {
	verbose, err := fs.GetBool(globalVerboseFlagName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	tfjson, err := fs.GetString(globalTerraformFlagName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &globalOptions{
		Verbose:        verbose,
		TerraformState: tfjson,
	}, nil
}

func initLogger(verbose bool) *logrus.Logger {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05 MST",
	}

	if verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	return logger
}

func loadClusterConfig(filename, terraformOutputPath string, logger *logrus.Logger) (*kubeoneapi.KubeOneCluster, error) {
	a, err := cluster.LoadKubeOneCluster(filename, terraformOutputPath, logger)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load a given KubeOneCluster object")
	}

	return a, nil
}
