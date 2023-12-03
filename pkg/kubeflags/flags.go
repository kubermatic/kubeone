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

	"k8c.io/kubeone/pkg/semverutil"
)

var (
	upToV126Constraint = semverutil.MustParseConstraint("< 1.26.0")
	v126Constraint     = semverutil.MustParseConstraint(">= 1.26.0, < 1.27.0")
)

// DefaultAdmissionControllers return list of default admission controllers for
// given kubernetes version
func DefaultAdmissionControllers(v *semver.Version) string {
	switch {
	case upToV126Constraint.Check(v):
		return strings.Join(defaultAdmissionControllersv1225, ",")
	case v126Constraint.Check(v):
		return strings.Join(defaultAdmissionControllersv1226, ",")
	default:
		return strings.Join(defaultAdmissionControllersv1227v1228, ",")
	}
}
