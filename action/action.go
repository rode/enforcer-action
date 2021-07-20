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
	"strings"

	"github.com/rode/evaluate-policy-action/config"
	rode "github.com/rode/rode/proto/v1alpha1"
	"go.uber.org/zap"
)

type EvaluateResourceAction struct {
	config *config.Config
	client rode.RodeClient
	logger *zap.Logger
}

type ActionResult struct {
	Pass             bool
	FailBuild        bool
	EvaluationReport string
}

func NewEvaluateResourceAction(logger *zap.Logger, conf *config.Config, client rode.RodeClient) *EvaluateResourceAction {
	return &EvaluateResourceAction{
		conf,
		client,
		logger,
	}
}

func (a *EvaluateResourceAction) Run(ctx context.Context) (*ActionResult, error) {
	a.logger.Info("Evaluating resource", zap.String("policyGroup", a.config.PolicyGroup), zap.String("resourceUri", a.config.ResourceUri))
	response, err := a.client.EvaluateResource(ctx, &rode.ResourceEvaluationRequest{
		PolicyGroup: a.config.PolicyGroup,
		ResourceUri: a.config.ResourceUri,
		Source: &rode.ResourceEvaluationSource{
			Name: "enforcer-action",
			Url:  fmt.Sprintf("%s/%s/actions/runs/%d", a.config.GitHub.GitHubServerUrl, a.config.GitHub.GitHubRepository, a.config.GitHub.GitHubRunId),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error evaluating resource: %s", err)
	}

	var b strings.Builder
	resultMessage := statusMessage(response.ResourceEvaluation.Pass)
	fmt.Fprintf(&b, "Resource %s evaluation (id: %s):\n", resultMessage, response.ResourceEvaluation.Id)

	for _, result := range response.PolicyEvaluations {
		policy, err := a.client.GetPolicy(ctx, &rode.GetPolicyRequest{Id: result.PolicyVersionId})
		if err != nil {
			return nil, err
		}

		fmt.Fprintln(&b, fmt.Sprintf("\npolicy \"%s\" %s", policy.Name, statusMessage(result.Pass)))

		for _, v := range result.Violations {
			fmt.Fprintln(&b, v.Message)
		}
	}

	return &ActionResult{
		FailBuild:        a.config.Enforce && !response.ResourceEvaluation.Pass,
		Pass:             response.ResourceEvaluation.Pass,
		EvaluationReport: b.String(),
	}, nil
}

func statusMessage(pass bool) string {
	if pass {
		return "PASSED"
	}
	return "FAILED"
}
