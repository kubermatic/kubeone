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
	constrainv122x = mustConstraint("1.22.x")
)

// DefaultAdmissionControllers return list of default admission controllers for
// given kubernetes version
func DefaultAdmissionControllers(kubeVersion *semver.Version) string {
	switch {
	case constrainv122x.Check(kubeVersion):
		return strings.Join(defaultAdmissionControllersv122x, ",")
	default:
		// return same as for last known release
		return strings.Join(defaultAdmissionControllersv122x, ",")
	}
}

func mustConstraint(c string) *semver.Constraints {
	constraint, err := semver.NewConstraint(c)
	if err != nil {
		panic(err)
	}

	return constraint
}
