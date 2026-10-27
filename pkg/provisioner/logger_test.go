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

package provisioner

import (
	"testing"

	"github.com/sirupsen/logrus"
)

// recordingHook records the entries written to a logrus logger.
type recordingHook struct {
	entries []*logrus.Entry
}

func (h *recordingHook) Levels() []logrus.Level { return logrus.AllLevels }

func (h *recordingHook) Fire(entry *logrus.Entry) error {
	h.entries = append(h.entries, entry)

	return nil
}

func TestMachineControllerLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(nopWriter{})
	hook := &recordingHook{}
	logger.AddHook(hook)

	log := newMachineControllerLogger(logger).With("machine", "demo-cp-0")
	log.Infow("Network configuration", "network", "", "subnetwork", "")
	log.Debug("debug message")
	log.Warn("warning message")
	log.Error("error message")

	want := []struct {
		level   logrus.Level
		message string
	}{
		{logrus.DebugLevel, "Network configuration"},
		{logrus.DebugLevel, "debug message"},
		{logrus.WarnLevel, "warning message"},
		{logrus.ErrorLevel, "error message"},
	}

	if len(hook.entries) != len(want) {
		t.Fatalf("got %d entries, want %d", len(hook.entries), len(want))
	}

	for i, w := range want {
		got := hook.entries[i]
		if got.Level != w.level || got.Message != w.message {
			t.Errorf("entry %d: got %s %q, want %s %q", i, got.Level, got.Message, w.level, w.message)
		}
		if got.Data["machine"] != "demo-cp-0" {
			t.Errorf("entry %d: logger fields not kept: %v", i, got.Data)
		}
	}

	if _, ok := hook.entries[0].Data["network"]; !ok {
		t.Errorf("entry fields not kept: %v", hook.entries[0].Data)
	}
}

func TestMachineControllerLoggerInfoHiddenByDefault(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(nopWriter{})
	hook := &recordingHook{}
	logger.AddHook(hook)

	// KubeOne logs at info level without --verbose
	newMachineControllerLogger(logger).Info("IP family doesn't specify dual stack")

	if len(hook.entries) != 0 {
		t.Errorf("informational machine-controller message shown without --verbose: %v", hook.entries[0].Message)
	}
}

type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }
