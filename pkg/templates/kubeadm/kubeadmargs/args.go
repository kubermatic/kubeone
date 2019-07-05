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

package kubeadmargs

// Args is a wrapper abstract type on top of kubeadm
type Args struct {
	APIServer    APIServer
	FeatureGates map[string]bool
}

// APIServer arguments
type APIServer struct {
	ExtraArgs map[string]string
}

// AppendMapStringStringExtraArg appends to CLI mapStringString additional flag
func (apiserver *APIServer) AppendMapStringStringExtraArg(k, v string) {
	value := v
	if originalValue, ok := apiserver.ExtraArgs[k]; ok {
		value = originalValue + "," + v
	}
	apiserver.ExtraArgs[k] = value
}

// New init empty Args
func New() *Args {
	return NewFrom(nil)
}

// NewFrom constructor for Args, based on provided apiserver extraargs
func NewFrom(apiServerExtraArgs map[string]string) *Args {
	if apiServerExtraArgs == nil {
		apiServerExtraArgs = map[string]string{}
	}

	return &Args{
		APIServer: APIServer{
			ExtraArgs: apiServerExtraArgs,
		},
		FeatureGates: map[string]bool{},
	}
}
