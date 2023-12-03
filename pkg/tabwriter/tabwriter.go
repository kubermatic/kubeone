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

package tabwriter

import (
	"io"

	"github.com/liggitt/tabwriter"
)

const (
	tabwriterMinWidth = 6
	tabwriterWidth    = 4
	tabwriterPadding  = 3
	tabwriterPadChar  = ' '
	tabwriterFlags    = tabwriter.RememberWidths
)

// New returns a tabwriter that translates tabbed columns in input into properly aligned text.
func New(output io.Writer) *tabwriter.Writer {
	return NewWithPadding(output, tabwriterPadding)
}

// New returns a tabwriter that translates tabbed columns in input into properly aligned text.
func NewWithPadding(output io.Writer, padding int) *tabwriter.Writer {
	return tabwriter.NewWriter(output, tabwriterMinWidth, tabwriterWidth, padding, tabwriterPadChar, tabwriterFlags)
}
