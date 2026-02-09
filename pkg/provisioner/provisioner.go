/*
Copyright 2025 The KubeOne Authors.

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

package provisioner

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	"k8c.io/machine-controller/pkg/cloudprovider"
	cloudprovidererrors "k8c.io/machine-controller/pkg/cloudprovider/errors"
	"k8c.io/machine-controller/pkg/cloudprovider/instance"
	cloudprovidertypes "k8c.io/machine-controller/pkg/cloudprovider/types"
	machinecontrollerlog "k8c.io/machine-controller/pkg/log"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"
	"k8c.io/machine-controller/sdk/providerconfig"
	"k8c.io/machine-controller/sdk/providerconfig/configvar"
)

const (
	maxRetrieForMachines = 5
	hostnameAnnotation   = "ssh-username"

	userDataTemplate = `#cloud-config
ssh_pwauth: false

{{- if .ProviderSpec.SSHPublicKeys }}
ssh_authorized_keys:
{{- range .ProviderSpec.SSHPublicKeys }}
- "{{ . }}"
{{- end }}
{{- end }}
`
)

func cleanupTemplateOutput(output string) (string, error) {
	// Valid YAML files are not allowed to have empty lines containing spaces or tabs.
	// So far only cleanup.
	woBlankLines := regexp.MustCompile(`(?m)^[ \t]+$`).ReplaceAllString(output, "")

	return woBlankLines, nil
}

func getUserData(pconfig *providerconfig.Config) (string, error) {
	data := struct {
		ProviderSpec *providerconfig.Config
	}{
		ProviderSpec: pconfig,
	}

	tmpl, err := template.New("user-data").Parse(userDataTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse user-data template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute user-data template: %w", err)
	}

	return cleanupTemplateOutput(buf.String())
}

type MachineInstance struct {
	inst    instance.Instance
	sshUser string
}

func CreateMachines(ctx context.Context, machines []clusterv1alpha1.Machine) ([]Machine, error) {
	providerData := &cloudprovidertypes.ProviderData{
		Ctx: ctx,
	}

	rawLog := machinecontrollerlog.New(false, machinecontrollerlog.FormatConsole)
	log := rawLog.Sugar()

	var instances []MachineInstance

	// TODO: Dump all the errors in an array and do the max that is possible without early exit
	for _, machine := range machines {
		prov, err := getProvider(ctx, machine)
		if err != nil {
			return nil, err
		}

		machineCreated := false
		providerInstance, err := prov.Get(ctx, log, &machine, providerData)
		if err != nil {
			// case 1: instance was not found and we are going to create one
			if errors.Is(err, cloudprovidererrors.ErrInstanceNotFound) {
				// Get userdata (needed to inject SSH keys to instances)
				pconfig, cfgErr := providerconfig.GetConfig(machine.Spec.ProviderSpec)
				if cfgErr != nil {
					return nil, fmt.Errorf("failed to get providerSpec: %w", cfgErr)
				}

				userdata, userdataErr := getUserData(pconfig)
				if userdataErr != nil {
					return nil, userdataErr
				}

				// Create the instance
				_, createErr := prov.Create(ctx, log, &machine, providerData, userdata)
				if createErr != nil {
					return nil, createErr
				}
				machineCreated = true
			} else if ok, _, _ := cloudprovidererrors.IsTerminalError(err); ok {
				// case 2: terminal error was returned and manual interaction is required to recover
				return nil, fmt.Errorf("failed to create machine at cloudprovider, due to %w", err)
			} else {
				// case 3: transient error was returned, requeue the request and try again in the future
				return nil, fmt.Errorf("failed to get instance from provider: %w", err)
			}
		}

		if machineCreated {
			for range maxRetrieForMachines {
				providerInstance, err = prov.Get(ctx, log, &machine, providerData)
				if err != nil {
					return nil, err
				}

				addresses := providerInstance.Addresses()
				if len(addresses) > 0 && publicAndPrivateIPExist(addresses) {
					break
				}

				time.Sleep(5 * time.Second)
			}
		}

		// Instance exists
		addresses := providerInstance.Addresses()
		if len(addresses) == 0 {
			return nil, fmt.Errorf("machine %s has not been assigned an IP yet", providerInstance.Name())
		}

		sshUser := "root"
		if user := machine.Annotations[hostnameAnnotation]; sshUser != "" {
			sshUser = user
		}

		machineInstance := MachineInstance{
			inst:    providerInstance,
			sshUser: sshUser,
		}

		instances = append(instances, machineInstance)
	}

	return getMachineProvisionerOutput(instances), nil
}

func getProvider(ctx context.Context, machine clusterv1alpha1.Machine) (cloudprovidertypes.Provider, error) {
	providerConfig, err := providerconfig.GetConfig(machine.Spec.ProviderSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider config: %w", err)
	}

	skg := configvar.NewResolver(ctx, nil)
	prov, err := cloudprovider.ForProvider(providerConfig.CloudProvider, skg)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloud provider %q: %w", providerConfig.CloudProvider, err)
	}

	return prov, nil
}
