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

	"github.com/google/go-github/v35/github"
	"github.com/rode/enforcer-action/config"
	rode "github.com/rode/rode/proto/v1alpha1"
	"go.uber.org/zap"
)

type EnforcerAction struct {
	config *config.Config
	client rode.RodeClient
	github *github.Client
	logger *zap.Logger
}

type ActionResult struct {
	Pass             bool
	FailBuild        bool
	EvaluationReport string
}

func NewEnforcerAction(logger *zap.Logger, conf *config.Config, client rode.RodeClient, githubClient *github.Client) *EnforcerAction {
	return &EnforcerAction{
		conf,
		client,
		githubClient,
		logger,
	}
}

func (a *EnforcerAction) Run(ctx context.Context) (*ActionResult, error) {
	a.logger.Info("Evaluating resource", zap.String("policyGroup", a.config.PolicyGroup), zap.String("resourceUri", a.config.ResourceUri))
	response, err := a.client.EvaluateResource(ctx, &rode.ResourceEvaluationRequest{
		PolicyGroup: a.config.PolicyGroup,
		ResourceUri: a.config.ResourceUri,
		Source: &rode.ResourceEvaluationSource{
			Name: "enforcer-action",
			Url:  fmt.Sprintf("%s/%s/actions/runs/%d", a.config.GitHub.ServerUrl, a.config.GitHub.Repository, a.config.GitHub.RunId),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error evaluating resource: %s", err)
	}

	report, err := a.createEvaluationReport(ctx, response)
	if err != nil {
		return nil, err
	}

	if a.config.GitHub.PullRequestNumber != 0 {
		a.logger.Info("Decorating pull request", zap.Int("pr", a.config.GitHub.PullRequestNumber))
		// repository here is the owner/repo slug provided by the environment variable GITHUB_REPOSITORY, which is set by default when running in GitHub Actions
		slug := strings.Split(a.config.GitHub.Repository, "/")
		org := slug[0]
		repo := slug[1]

		// use the issues API to post a comment that's not attached to a line in the pull request diff
		// see https://docs.github.com/en/rest/reference/pulls#create-a-review-comment-for-a-pull-request
		_, _, err = a.github.Issues.CreateComment(ctx, org, repo, a.config.GitHub.PullRequestNumber, &github.IssueComment{
			Body: github.String(report),
		})

		if err != nil {
			return nil, fmt.Errorf("error posting pull request comment: %s", err)
		}
	}

	return &ActionResult{
		FailBuild:        a.config.Enforce && !response.ResourceEvaluation.Pass,
		Pass:             response.ResourceEvaluation.Pass,
		EvaluationReport: report,
	}, nil
}

func (a *EnforcerAction) createEvaluationReport(ctx context.Context, evaluationResult *rode.ResourceEvaluationResult) (string, error) {
	resourceEval := evaluationResult.ResourceEvaluation
	md := markdownPrinter{}
	md.
		h1("Rode Resource Evaluation Report %s", statusMessage(resourceEval.Pass)).
		quote("report id: "+resourceEval.Id).
		h2("Resource Metadata").
		table([]string{"Resource URI"}, [][]string{
			{asCode(resourceEval.ResourceVersion.Version)},
		})

	if len(resourceEval.ResourceVersion.Names) > 0 {
		var artifactNames []string
		for _, name := range resourceEval.ResourceVersion.Names {
			artifactNames = append(artifactNames, asCode(name))
		}

		md.h3("Artifact Names").list(artifactNames).newline()
	}

	md.h2("Policy Results")
	for _, result := range evaluationResult.PolicyEvaluations {
		policy, err := a.client.GetPolicy(ctx, &rode.GetPolicyRequest{Id: result.PolicyVersionId})
		if err != nil {
			return "", err
		}

		md.
			h3("%s %s", policy.Name, statusMessage(result.Pass)).
			codeBlock()

		for _, v := range result.Violations {
			md.write(v.Message)
		}
		md.codeBlock().newline()
	}

	return md.string(), nil
}

func statusMessage(pass bool) string {
	if pass {
		return "✅ (PASSED)"
	}

	return "❌ (FAILED)"
}