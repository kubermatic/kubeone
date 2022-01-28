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
	"bufio"
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"golang.org/x/term"

	"k8c.io/kubeone/pkg/addons"
	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	"k8c.io/kubeone/pkg/apis/kubeone/config"
	"k8c.io/kubeone/pkg/credentials"
	"k8c.io/kubeone/pkg/state"
)

const yes = "yes"

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

	// Validate Addons path if provided
	if s.Cluster.Addons.Enabled() {
		addonsPath, err := s.Cluster.Addons.RelativePath(s.ManifestFilePath)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get addons path")
		}

		// Check if only embedded addons are being used; path is not required for embedded addons and no validation is required
		embeddedAddonsOnly, err := addons.EmbeddedAddonsOnly(s.Cluster.Addons.Addons)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read embedded addons directory")
		}
		// If custom addons are being used then addons path is required and should be a valid directory
		if !embeddedAddonsOnly {
			if _, err := os.Stat(addonsPath); os.IsNotExist(err) {
				return nil, errors.Wrapf(err, "failed to validate addons path, make sure that directory %q exists", s.Cluster.Addons.Path)
			}
		}
	}

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

func confirmCommand(autoApprove bool) (bool, error) {
	if autoApprove {
		return true, nil
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return false, errors.New("not running in the terminal")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to proceed (yes/no): ")

	confirmation, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	fmt.Println()

	return strings.Trim(confirmation, "\n") == yes, nil
}

func validateCredentials(s *state.State, credentialsFile string) error {
	_, universalErr := credentials.ProviderCredentials(s.Cluster.CloudProvider, credentialsFile, credentials.TypeUniversal)
	_, mcErr := credentials.ProviderCredentials(s.Cluster.CloudProvider, credentialsFile, credentials.TypeMC)
	_, ccmErr := credentials.ProviderCredentials(s.Cluster.CloudProvider, credentialsFile, credentials.TypeCCM)

	switch {
	case universalErr != nil && mcErr != nil && ccmErr != nil:
		// No credentials found
		fallthrough
	case mcErr == nil && ccmErr != nil && universalErr != nil:
		// MC credentials found, but no CCM or universal credentials
		fallthrough
	case ccmErr == nil && mcErr != nil && universalErr != nil: // CCM credentials found, but no MC or universal credentials
		return errors.Wrap(universalErr, "failed to validate credentials")
	default:
		return nil
	}
}
