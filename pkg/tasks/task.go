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
	"time"

	"k8c.io/kubeone/pkg/state"

	"k8s.io/apimachinery/pkg/util/wait"
)

// defaultRetryBackoff is backoff with with duration of 5 seconds and factor of 2.0
func defaultRetryBackoff(retries int) wait.Backoff {
	return wait.Backoff{
		Steps:    retries,
		Duration: 5 * time.Second,
		Factor:   2.0,
	}
}

// Task is a runnable task
type Task struct {
	Fn          func(*state.State) error
	Predicate   func(*state.State) bool
	Description string
	ErrMsg      string
	Retries     int
}

// Run runs a task
func (t *Task) Run(s *state.State) error {
	if t.Retries == 0 {
		t.Retries = 10
	}

	backoff := defaultRetryBackoff(t.Retries)

	var lastError error
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		if lastError != nil {
			s.Logger.Warn("Retrying task...")
		}

		lastError = t.Fn(s)
		if lastError != nil {
			s.Logger.Warnf("Task failed, error was: %s", lastError)
			return false, nil
		}
		return true, nil
	})
	if err == wait.ErrWaitTimeout {
		err = lastError
	}
	return err
}
