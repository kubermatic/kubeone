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

package kubeflags

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

var (
	constrainv114x      = mustConstraint(">= 1.14.0, < 1.15.0")
	constrainv115x      = mustConstraint("1.15.x")
	constrainv116xv117x = mustConstraint(">= 1.16.0, < 1.18.0")
	constrainv118xv121x = mustConstraint(">= 1.18.0, < 1.22.0")
	constrainv122x      = mustConstraint("1.22.x")
)

// DefaultAdmissionControllers return list of default admission controllers for
// given kubernetes version
func DefaultAdmissionControllers(kubeVersion *semver.Version) string {
	switch {
	case constrainv114x.Check(kubeVersion):
		return strings.Join(defaultAdmissionControllersv114x, ",")
	case constrainv115x.Check(kubeVersion):
		return strings.Join(defaultAdmissionControllersv115x, ",")
	case constrainv116xv117x.Check(kubeVersion):
		return strings.Join(defaultAdmissionControllersv116xv117x, ",")
	case constrainv118xv121x.Check(kubeVersion):
		return strings.Join(defaultAdmissionControllersv118xv121x, ",")
	case constrainv122x.Check(kubeVersion):
		return strings.Join(defaultAdmissionControllersv122x, ",")
	}

	// return same as for last known release
	return strings.Join(defaultAdmissionControllersv122x, ",")
}

func mustConstraint(c string) *semver.Constraints {
	constraint, err := semver.NewConstraint(c)
	if err != nil {
		panic(err)
	}

	return constraint
}
