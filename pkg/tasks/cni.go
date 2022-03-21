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

package tasks

import (
	"github.com/pkg/errors"

	"k8c.io/kubeone/pkg/fail"
	"k8c.io/kubeone/pkg/state"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ensureCNI(s *state.State) error {
	var err error
	if s.Cluster.ClusterNetwork.CNI.External != nil {
		s.Logger.Infoln("External CNI plugin will be used")
	}

	if s.RESTConfig == nil {
		return fail.KubeClientError{
			Op:  "ensure CNI",
			Err: errors.New("rest config is not initialized"),
		}
	}

	s.DynamicClient, err = client.New(s.RESTConfig, client.Options{})
	if err != nil {
		return fail.KubeClient(err, "create kubernetes dynamic client")
	}

	return nil
}
