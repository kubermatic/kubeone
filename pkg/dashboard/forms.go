/*
Copyright 2026 The KubeOne Authors.

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

package dashboard

import (
	"net/http"

	forms "github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
)

func parseForm[T any](req *http.Request) (T, error) {
	var val T

	if err := req.ParseForm(); err != nil {
		return val, err
	}

	return val, forms.NewDecoder().Decode(&val, req.PostForm)
}

func parseAndValidateForm[T any](req *http.Request) (T, error) {
	val, err := parseForm[T](req)
	if err != nil {
		return val, err
	}

	return val, validator.New().Struct(val)
}

type namespaceNameForm struct {
	Namespace string `validate:"required"`
	Name      string `validate:"required"`
}

type scaleForm struct {
	namespaceNameForm
	Direction string `validate:"required,oneof=up down"`
}
