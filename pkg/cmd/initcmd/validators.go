/*
Copyright 2022 The KubeOne Authors.

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

package initcmd

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	kubeoneapi "k8c.io/kubeone/pkg/apis/kubeone"
	kubeonevalidation "k8c.io/kubeone/pkg/apis/kubeone/validation"
)

func clusterNameValidator(val interface{}) error {
	if str, ok := val.(string); ok {
		if errs := kubeonevalidation.ValidateName(str, nil); len(errs) > 0 {
			return fmt.Errorf("provided value is not a valid cluster name: %w", errs.ToAggregate())
		}
	} else {
		return fmt.Errorf("cluster name must be a valid string, but got %v", reflect.TypeOf(val).Name())
	}

	return nil
}

func kubernetesVersionValidator(val interface{}) error {
	if str, ok := val.(string); ok {
		if errs := kubeonevalidation.ValidateVersionConfig(kubeoneapi.VersionConfig{Kubernetes: strings.TrimLeft(str, "v")}, nil); len(errs) > 0 {
			return fmt.Errorf("provided value is not a valid kubernetes version: %w", errs.ToAggregate())
		}
	} else {
		return fmt.Errorf("kubernetes version must be a valid semver, but got %v", reflect.TypeOf(val).Name())
	}

	return nil
}

func positiveNumberValidator(val interface{}) error {
	if i, ok := val.(int); ok {
		if i <= 0 {
			return fmt.Errorf("provided value must be positive, but got %d", i)
		}

		return nil
	}
	if str, ok := val.(string); ok {
		i, err := strconv.Atoi(str)
		if err != nil {
			return fmt.Errorf("provided value must be a valid integer, but got %v", reflect.TypeOf(val).Name())
		}
		if i <= 0 {
			return fmt.Errorf("provided value must be positive, but got %d", i)
		}
	}

	return nil
}

func oddNumberValidator(val interface{}) error {
	if i, ok := val.(int); ok {
		if i <= 0 {
			return fmt.Errorf("provided value must be positive, but got %d", i)
		}
		if i%2 == 0 {
			return fmt.Errorf("provided value must be odd, but got %d", i)
		}

		return nil
	}
	if str, ok := val.(string); ok {
		i, err := strconv.Atoi(str)
		if err != nil {
			return fmt.Errorf("provided value must be a valid integer, but got %v", reflect.TypeOf(val).Name())
		}
		if i <= 0 {
			return fmt.Errorf("provided value must be positive, but got %d", i)
		}
		if i%2 == 0 {
			return fmt.Errorf("provided value must be odd, but got %d", i)
		}
	}

	return nil
}
