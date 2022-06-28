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
	"path/filepath"
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
	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"
)

const yes = "yes"

type globalOptions struct {
	ManifestFile    string `longflag:"manifest" shortflag:"m"`
	TerraformState  string `longflag:"tfjson" shortflag:"t"`
	CredentialsFile string `longflag:"credentials" shortflag:"c"`
	Verbose         bool   `longflag:"verbose" shortflag:"v"`
	Debug           bool   `longflag:"debug" shortflag:"d"`
	LogFormat       string `longflag:"log-format" shortflag:"l"`
}

func (opts *globalOptions) BuildState() (*state.State, error) {
	rootContext := context.Background()
	s, err := state.New(rootContext)
	if err != nil {
		return nil, err
	}

	s.Logger = newLogger(opts.Verbose, opts.LogFormat)

	cluster, err := loadClusterConfig(opts.ManifestFile, opts.TerraformState, opts.CredentialsFile, s.Logger)
	if err != nil {
		return nil, err
	}

	s.Cluster = cluster
	s.ManifestFilePath = opts.ManifestFile
	s.CredentialsFilePath = opts.CredentialsFile
	s.Verbose = opts.Verbose

	// Validate Addons path if provided
	if s.Cluster.Addons.Enabled() {
		addonsPath, err := s.Cluster.Addons.RelativePath(s.ManifestFilePath)
		if err != nil {
			return nil, err
		}

		// Check if only embedded addons are being used; path is not required for embedded addons and no validation is required
		embeddedAddonsOnly, err := addons.EmbeddedAddonsOnly(s.Cluster.Addons.Addons)
		if err != nil {
			return nil, err
		}

		// If custom addons are being used then addons path is required and should be a valid directory
		if !embeddedAddonsOnly {
			if _, err := os.Stat(addonsPath); os.IsNotExist(err) {
				return nil, fail.Runtime(err, "checking addons directory")
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
		return nil, fail.Runtime(err, "getting global flags")
	}
	gf.ManifestFile = manifestFile

	verbose, err := fs.GetBool(longFlagName(gf, "Verbose"))
	if err != nil {
		return nil, fail.Runtime(err, "getting global flags")
	}
	gf.Verbose = verbose

	tfjson, err := fs.GetString(longFlagName(gf, "TerraformState"))
	if err != nil {
		return nil, fail.Runtime(err, "getting global flags")
	}
	gf.TerraformState = tfjson

	creds, err := fs.GetString(longFlagName(gf, "CredentialsFile"))
	if err != nil {
		return nil, fail.Runtime(err, "getting global flags")
	}
	gf.CredentialsFile = creds

	logFormat, err := fs.GetString(longFlagName(gf, "LogFormat"))
	if err != nil {
		return nil, fail.Runtime(err, "getting global flags")
	}
	gf.LogFormat = logFormat

	return gf, nil
}

func newLogger(verbose bool, format string) *logrus.Logger {
	logger := logrus.New()

	switch format {
	case "json":
		logger.Formatter = &logrus.JSONFormatter{}
	default:
		logger.Formatter = &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "15:04:05 MST",
		}
	}

	if verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	return logger
}

func loadClusterConfig(filename, terraformOutputPath, credentialsFilePath string, logger logrus.FieldLogger) (*kubeoneapi.KubeOneCluster, error) {
	cls, err := config.LoadKubeOneCluster(filename, terraformOutputPath, credentialsFilePath, logger)
	if err != nil {
		return nil, err
	}

	return cls, nil
}

func confirmCommand(autoApprove bool) (bool, error) {
	if autoApprove {
		return true, nil
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return false, fail.Runtime(fmt.Errorf("not running in the terminal"), "terminal detecting")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to proceed (yes/no): ")

	confirmation, err := reader.ReadString('\n')
	if err != nil {
		return false, fail.Runtime(err, "reading confirmation")
	}

	fmt.Println()

	return strings.Trim(confirmation, "\n") == yes, nil
}

func validateCredentials(s *state.State, credentialsFile string) error {
	_, universalErr := credentials.ProviderCredentials(s.Cluster.CloudProvider, credentialsFile, credentials.TypeUniversal)

	var mcErr error
	if s.Cluster.MachineController.Deploy {
		_, mcErr = credentials.ProviderCredentials(s.Cluster.CloudProvider, credentialsFile, credentials.TypeMC)
	}

	_, ccmErr := credentials.ProviderCredentials(s.Cluster.CloudProvider, credentialsFile, credentials.TypeCCM)

	switch {
	case universalErr != nil && mcErr != nil && ccmErr != nil:
		// No credentials found
		fallthrough
	case mcErr == nil && ccmErr != nil && universalErr != nil:
		// MC credentials found, but no CCM or universal credentials
		fallthrough
	case ccmErr == nil && mcErr != nil && universalErr != nil: // CCM credentials found, but no MC or universal credentials
		return fail.ConfigValidation(universalErr)
	default:
		return nil
	}
}

func initBackup(backupPath string) error {
	// refuse to overwrite existing backups (NB: since we attempt to
	// write to the file later on to check for write permissions, we
	// inadvertently create a zero byte file even if the first step
	// of the installer fails; for this reason it's okay to find an
	// existing, zero byte backup)
	fi, err := os.Stat(backupPath)
	if err != nil && fi != nil && fi.Size() > 0 {
		return fail.RuntimeError{
			Op:  fmt.Sprintf("checking backup file %s existence", backupPath),
			Err: errors.New("refusing to overwrite"),
		}
	}

	// try to write to the file before doing anything else
	backup, err := os.OpenFile(backupPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fail.Runtime(err, "opening backup file for write")
	}

	return fail.Runtime(backup.Close(), "closing backup file")
}

func defaultBackupPath(backupPath, manifestPath, clusterName string) string {
	if backupPath == "" {
		fullPath, _ := filepath.Abs(manifestPath)
		backupPath = filepath.Join(filepath.Dir(fullPath), fmt.Sprintf("%s.tar.gz", clusterName))
	}

	return backupPath
}
