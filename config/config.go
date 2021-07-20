// Copyright 2021 The Rode Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"errors"
	"flag"
	"strings"

	"github.com/peterbourgon/ff/v3"
	"github.com/rode/rode/common"
)

type GitHubConfig struct {
	GitHubRunId      int
	GitHubServerUrl  string
	GitHubRepository string
}

type Config struct {
	AccessToken  string
	GitHub       *GitHubConfig
	Enforce      bool
	PolicyGroup  string
	ResourceUri  string
	ClientConfig *common.ClientConfig
}

func Build(name string, args []string) (*Config, error) {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	c := &Config{
		ClientConfig: common.SetupRodeClientFlags(flags),
		GitHub:       &GitHubConfig{},
	}

	flags.StringVar(&c.AccessToken, "access-token", "", "An access token that will be included in requests to Rode.")
	flags.BoolVar(&c.Enforce, "enforce", true, "Controls whether the step should fail if the evaluation fails.")
	flags.StringVar(&c.PolicyGroup, "policy-group", "", "The policy group to evaluate the resource against.")
	flags.StringVar(&c.ResourceUri, "resource-uri", "", "The resource to evaluate policy against.")
	flags.StringVar(&c.GitHub.GitHubServerUrl, "github-server-url", "", "The GitHub server url. This is set automatically when running in GitHub Actions.")
	flags.StringVar(&c.GitHub.GitHubRepository, "github-repository", "", "An org/repo slug. This is set automatically when running in GitHub Actions.")
	flags.IntVar(&c.GitHub.GitHubRunId, "github-run-id", 0, "The run id of a workflow. This is set automatically when running in GitHub Actions.")

	if err := ff.Parse(flags, args, ff.WithEnvVarNoPrefix()); err != nil {
		return nil, err
	}

	c.PolicyGroup = strings.TrimSpace(c.PolicyGroup)

	if c.PolicyGroup == "" {
		return nil, errors.New("must set policy-group")
	}

	if c.ResourceUri == "" {
		return nil, errors.New("must set resource-uri")
	}

	return c, nil
}
