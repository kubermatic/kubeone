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
	"context"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	kubeoneapi "github.com/kubermatic/kubeone/pkg/apis/kubeone"
	"github.com/kubermatic/kubeone/pkg/apis/kubeone/config"
	"github.com/kubermatic/kubeone/pkg/state"
)

type globalOptions struct {
	ManifestFile    string `longflag:"manifest" shortflag:"m"`
	TerraformState  string `longflag:"tfjson" shortflag:"t"`
	CredentialsFile string `longflag:"credentials" shortflag:"c"`
	Verbose         bool   `longflag:"verbose" shortflag:"v"`
	Debug           bool   `longflag:"debug" shortflag:"d"`
}

func (opts *globalOptions) BuildState() (*state.State, error) {
	rootContext := context.Background()
	s, err := state.New(rootContext)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize State")
	}
	s.Logger = newLogger(opts.Verbose)

	cluster, err := loadClusterConfig(opts.ManifestFile, opts.TerraformState, opts.CredentialsFile, s.Logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load cluster")
	}

	s.Cluster = cluster
	s.ManifestFilePath = opts.ManifestFile
	s.CredentialsFilePath = opts.CredentialsFile
	s.Verbose = opts.Verbose

	return s, nil
}

func longFlagName(obj interface{}, fieldName string) string {
	elem := reflect.TypeOf(obj).Elem()
	field, ok := elem.FieldByName(fieldName)
	if !ok {
		return strings.ToLower(fieldName)
	}

	return field.Tag.Get("longflag")
}

func shortFlagName(obj interface{}, fieldName string) string {
	elem := reflect.TypeOf(obj).Elem()
	field, _ := elem.FieldByName(fieldName)
	return field.Tag.Get("shortflag")
}

func persistentGlobalOptions(fs *pflag.FlagSet) (*globalOptions, error) {
	gf := &globalOptions{}

	manifestFile, err := fs.GetString(longFlagName(gf, "ManifestFile"))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	gf.ManifestFile = manifestFile

	verbose, err := fs.GetBool(longFlagName(gf, "Verbose"))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	gf.Verbose = verbose

	tfjson, err := fs.GetString(longFlagName(gf, "TerraformState"))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	gf.TerraformState = tfjson

	creds, err := fs.GetString(longFlagName(gf, "CredentialsFile"))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	gf.CredentialsFile = creds

	return gf, nil
}

func newLogger(verbose bool) *logrus.Logger {
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

func loadClusterConfig(filename, terraformOutputPath, credentialsFilePath string, logger logrus.FieldLogger) (*kubeoneapi.KubeOneCluster, error) {
	a, err := config.LoadKubeOneCluster(filename, terraformOutputPath, credentialsFilePath, logger)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load a given KubeOneCluster object")
	}

	return a, nil
}
