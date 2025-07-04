/*
Copyright 2024 The KubeOne Authors.

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

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/go-github/v65/github"

	"k8c.io/kubeone/test/e2e"

	"k8s.io/utils/ptr"
)

const GitHubTokenEnv = "GITHUB_TOKEN" //nolint:gosec

var providerNameMap = map[string]string{
	"aws":          "AWS",
	"azure":        "Azure",
	"digitalocean": "DigitalOcean",
	"equinixmetal": "Equinix Metal",
	"gce":          "GCE",
	"hetzner":      "Hetzner",
	"openstack":    "OpenStack",
	"vsphere":      "vSphere",
}

type GitHubIssue struct {
	Cases map[string]GitHubIssueCase
}

type GitHubIssueCase struct {
	TestCommands map[string][]string
}

func generateGitHubIssues(outputBuf io.ReadWriter, kubeoneTest []KubeoneTest, providerName, scenarioName, kubeoneVersion string, createIssues bool) error {
	if createIssues {
		if os.Getenv(GitHubTokenEnv) == "" {
			return errors.New("cannot file testing issue as GitHub token is not set")
		}
	}

	issues := map[string]GitHubIssue{}

	for _, kt := range kubeoneTest {
		if issues[kt.Scenario].Cases == nil {
			issues[kt.Scenario] = GitHubIssue{
				Cases: map[string]GitHubIssueCase{},
			}
		}

		if scenarioName != "" && kt.Scenario != scenarioName {
			continue
		}

		for _, i := range kt.Infrastructures {
			infra := e2e.Infrastructures[i.Name]

			if (providerName != "" && infra.Provider() != providerName) || infra.DiscludeFromIssue() {
				continue
			}

			if issues[kt.Scenario].Cases[infra.Provider()].TestCommands == nil {
				issues[kt.Scenario].Cases[infra.Provider()] = GitHubIssueCase{
					TestCommands: map[string][]string{},
				}
			}

			if issues[kt.Scenario].Cases[infra.Provider()].TestCommands[infra.OperatingSystem()] == nil {
				issues[kt.Scenario].Cases[infra.Provider()].TestCommands[infra.OperatingSystem()] = []string{}
			}

			var testCommand string
			if kt.UpgradedVersion != "" {
				testCommand = e2e.PullProwJobName(i.Name, kt.Scenario, "from", kt.InitVersion, "to", kt.UpgradedVersion)
			} else {
				testCommand = e2e.PullProwJobName(i.Name, kt.Scenario, kt.InitVersion)
			}

			issues[kt.Scenario].Cases[infra.Provider()].TestCommands[infra.OperatingSystem()] = append(
				issues[kt.Scenario].Cases[infra.Provider()].TestCommands[infra.OperatingSystem()],
				fmt.Sprintf("/test %s", testCommand),
			)
		}
	}

	for scenario, tests := range issues {
		scenarioName := e2e.Scenarios[scenario].GetHumanReadableName()
		for provider, cases := range tests.Cases {
			tpl, tErr := template.New("").Parse(githubIssueTemplate)
			if tErr != nil {
				return tErr
			}

			var out bytes.Buffer
			if err := tpl.Execute(&out, cases); err != nil {
				return err
			}

			title := fmt.Sprintf("Test KubeOne Release %s - %s on %s", kubeoneVersion, scenarioName, providerNameMap[provider])

			if createIssues {
				gh := github.NewClient(nil).WithAuthToken(os.Getenv(GitHubTokenEnv))

				body := out.String()
				request := &github.IssueRequest{
					Title: &title,
					Body:  &body,
					Labels: &[]string{
						"sig/cluster-management",
						"priority/high",
					},
					State: ptr.To("open"),
				}

				issue, _, err := gh.Issues.Create(context.Background(), "kubermatic", "kubeone", request)
				if err != nil {
					return fmt.Errorf("creating testing issue: %w", err)
				}

				fmt.Fprintf(outputBuf, "Testing issue created #%d (https://github.com/kubermatic/kubeone/issue/%d)!\n", issue.GetNumber(), issue.GetNumber())
			} else {
				fmt.Fprintf(outputBuf, "%s\n\n", title)
				if _, err := outputBuf.Write(out.Bytes()); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

var githubIssueTemplate = heredoc.Doc(`
This is a testing ticket for the upcoming KubeOne release.

_This issue has been automatically generated._

{{ range $os, $cmds := .TestCommands }}
### {{ $os }}
{{- range $cmds }}
- [ ] ` + "`" + `{{ . }}` + "`" + `{{ "" -}}
{{ end }}
{{ end }}
`)
