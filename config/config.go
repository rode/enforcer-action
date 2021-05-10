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
	"context"
	"errors"
	"strings"

	"github.com/sethvargo/go-envconfig"
)

type RodeConfig struct {
	Host     string `env:"HOST,required"`
	Insecure bool   `env:"INSECURE,required"`
}

type Config struct {
	Enforce     bool        `env:"ENFORCE,required"`
	PolicyId    string      `env:"POLICY_ID"`
	PolicyName  string      `env:"POLICY_NAME"`
	ResourceUri string      `env:"RESOURCE_URI,required"`
	Rode        *RodeConfig `env:",prefix=RODE_"`
}

func Build(ctx context.Context, l envconfig.Lookuper) (*Config, error) {
	c := &Config{}

	if err := envconfig.ProcessWith(ctx, c, l); err != nil {
		return nil, err
	}

	c.PolicyName = strings.TrimSpace(c.PolicyName)
	c.PolicyId = strings.TrimSpace(c.PolicyId)

	if c.PolicyId == "" && c.PolicyName == "" {
		return nil, errors.New("must set either policyName or policyId")
	}

	if c.PolicyId != "" && c.PolicyName != "" {
		return nil, errors.New("only one of policyId or policyName should be specified")
	}

	return c, nil
}
