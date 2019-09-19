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

	"github.com/Masterminds/semver"
)

var (
	constrainv1130v1132 = mustConstraint(">= 1.13.0, < 1.13.3")
	constrainv1133v114x = mustConstraint(">= 1.13.3, < 1.15.0")
	constrainv115x      = mustConstraint("1.15.x")
	constrainv116x      = mustConstraint("1.16.x")
)

// DefaultAdmissionControllers return list of default admission controllers for
// given kubernetes version
func DefaultAdmissionControllers(kubeVersion *semver.Version) string {
	switch {
	case constrainv1130v1132.Check(kubeVersion):
		return strings.Join(defaultAdmissionControllersv1130v1132, ",")
	case constrainv1133v114x.Check(kubeVersion):
		return strings.Join(defaultAdmissionControllersv1133v114x, ",")
	case constrainv115x.Check(kubeVersion):
		return strings.Join(defaultAdmissionControllersv115x, ",")
	case constrainv116x.Check(kubeVersion):
		return strings.Join(defaultAdmissionControllersv116x, ",")
	}

	// return same as for last known release
	return strings.Join(defaultAdmissionControllersv116x, ",")
}

func mustConstraint(c string) *semver.Constraints {
	constraint, err := semver.NewConstraint(c)
	if err != nil {
		panic(err)
	}

	return constraint
}
