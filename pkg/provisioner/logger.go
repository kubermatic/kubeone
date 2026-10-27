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
	"slices"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newMachineControllerLogger returns a logger for the machine-controller
// cloud provider drivers, which KubeOne runs in-process to manage control
// plane machines. Their logs are written to the KubeOne logger, so they follow
// its format and verbosity: informational messages (e.g. the network
// configuration of a new instance) are debug messages, shown only with
// --verbose, warnings and errors are always shown.
func newMachineControllerLogger(logger logrus.FieldLogger) *zap.SugaredLogger {
	return zap.New(&logrusCore{logger: logger}).Sugar()
}

// logrusCore is a zapcore.Core writing the entries to a logrus logger.
type logrusCore struct {
	logger logrus.FieldLogger
	fields []zapcore.Field
}

var _ zapcore.Core = &logrusCore{}

func (c *logrusCore) Enabled(zapcore.Level) bool {
	return true
}

func (c *logrusCore) With(fields []zapcore.Field) zapcore.Core {
	return &logrusCore{
		logger: c.logger,
		fields: append(slices.Clone(c.fields), fields...),
	}
}

func (c *logrusCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return checked.AddCore(entry, c)
}

func (c *logrusCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	encoder := zapcore.NewMapObjectEncoder()
	for _, field := range c.fields {
		field.AddTo(encoder)
	}
	for _, field := range fields {
		field.AddTo(encoder)
	}

	logger := c.logger
	if len(encoder.Fields) > 0 {
		logger = logger.WithFields(logrus.Fields(encoder.Fields))
	}

	switch {
	case entry.Level >= zapcore.ErrorLevel:
		logger.Error(entry.Message)
	case entry.Level == zapcore.WarnLevel:
		logger.Warn(entry.Message)
	default:
		logger.Debug(entry.Message)
	}

	return nil
}

func (c *logrusCore) Sync() error {
	return nil
}
