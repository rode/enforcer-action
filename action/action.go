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

package action

import (
	"context"
	"fmt"

	"github.com/rode/evaluate-policy-action/config"
	rode "github.com/rode/rode/proto/v1alpha1"
	"go.uber.org/zap"
)

type EvaluatePolicyAction struct {
	config *config.Config
	client rode.RodeClient
	logger *zap.Logger
}

func NewEvaluatePolicyAction(logger *zap.Logger, conf *config.Config, client rode.RodeClient) *EvaluatePolicyAction {
	return &EvaluatePolicyAction{
		conf,
		client,
		logger,
	}
}

func (a *EvaluatePolicyAction) Run(ctx context.Context) (bool, error) {
	policyId, err := a.determinePolicyId(ctx)
	if err != nil {
		return false, err
	}

	a.logger.Info("Evaluating policy", zap.String("policyId", policyId))
	response, err := a.client.EvaluatePolicy(ctx, &rode.EvaluatePolicyRequest{
		Policy:      policyId,
		ResourceUri: a.config.ResourceUri,
	})

	if err != nil {
		return false, fmt.Errorf("error evaluating policy: %s", err)
	}

	return response.Pass, nil
}

func (a *EvaluatePolicyAction) determinePolicyId(ctx context.Context) (string, error) {
	if a.config.PolicyId != "" {
		a.logger.Info("Taking policy id from config", zap.String("policyId", a.config.PolicyId))
		return a.config.PolicyId, nil
	}

	a.logger.Info("Fetching policy by name", zap.String("policyName", a.config.PolicyName))
	policies, err := a.client.ListPolicies(ctx, &rode.ListPoliciesRequest{
		Filter: fmt.Sprintf(`name == "%s"`, a.config.PolicyName),
	})

	if err != nil {
		return "", fmt.Errorf("error fetching policies: %s", err)
	}

	if len(policies.Policies) == 0 {
		return "", fmt.Errorf("unable to find a policy for %s", a.config.PolicyName)
	}

	if len(policies.Policies) > 1 {
		a.logger.Warn("Found more than one policy matching name, taking the first", zap.String("name", a.config.PolicyName))
	}

	return policies.Policies[0].Id, nil
}
