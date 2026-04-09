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

package confirmation

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"k8c.io/kubeone/pkg/fail"
)

const yes = "yes"

func Approved(autoApprove bool) (bool, error) {
	if autoApprove {
		return true, nil
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return false, fail.Runtime(fmt.Errorf("not running in the terminal"), "terminal detecting")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to proceed (yes/no): ")

	confirmation, err := reader.ReadString('\n')
	if err != nil {
		return false, fail.Runtime(err, "reading confirmation")
	}

	fmt.Println()

	return strings.Trim(confirmation, "\n") == yes, nil
}
