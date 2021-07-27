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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v35/github"
	"github.com/rode/enforcer-action/config"
	rode "github.com/rode/rode/proto/v1alpha1"
	"go.uber.org/zap"
)

const (
	githubPrEventName                 = "pull_request"
	githubPrTargetEventName           = "pull_request_target"
	evaluationReportCommentIdentifier = "generated-by: enforcer-action"
)

var osReadFile = os.ReadFile

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

type pullRequest struct {
	Number int `json:"number"`
}

type pullRequestEvent struct {
	PullRequest *pullRequest `json:"pull_request"`
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

	if err = a.decoratePullRequest(ctx, report); err != nil {
		return nil, err
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

	// leave an HTML comment in the markdown so that we can find the comment on future job runs
	md.comment(evaluationReportCommentIdentifier)

	return md.string(), nil
}

func (a *EnforcerAction) decoratePullRequest(ctx context.Context, comment string) error {
	shouldDecoratePullRequest := (a.config.GitHub.EventName == githubPrEventName ||
		a.config.GitHub.EventName == githubPrTargetEventName) && a.config.GitHub.EventPath != ""

	if !shouldDecoratePullRequest {
		a.logger.Info("Skipping pull request decoration")
		return nil
	}

	eventJson, err := osReadFile(a.config.GitHub.EventPath)
	if err != nil {
		return fmt.Errorf("error reading event payload at %s: %s", a.config.GitHub.EventPath, err)
	}

	var prEvent pullRequestEvent
	if err := json.Unmarshal(eventJson, &prEvent); err != nil {
		return fmt.Errorf("error unmarshalling event json: %s", err)
	}

	a.logger.Info("Decorating pull request", zap.Int("pr", prEvent.PullRequest.Number))
	// repository here is the owner/repo slug provided by the environment variable GITHUB_REPOSITORY, which is set by default when running in GitHub Actions
	slug := strings.Split(a.config.GitHub.Repository, "/")
	org := slug[0]
	repo := slug[1]

	var existingCommentId int64
	comments, _, err := a.github.Issues.ListComments(ctx, org, repo, prEvent.PullRequest.Number, &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})

	if err != nil {
		return fmt.Errorf("error search for existing pull request comment: %s", err)
	}

	for _, prComment := range comments {
		if strings.Contains(*prComment.Body, evaluationReportCommentIdentifier) {
			existingCommentId = *prComment.ID
			break
		}
	}

	if existingCommentId != 0 {
		a.logger.Info("Found existing comment, updating", zap.Int64("commentId", existingCommentId))
		_, _, err := a.github.Issues.EditComment(ctx, org, repo, existingCommentId, &github.IssueComment{
			Body: github.String(comment),
		})

		if err != nil {
			return fmt.Errorf("error updating comment (id: %d): %s", existingCommentId, err)
		}

		return nil
	}

	// use the issues API to post a comment that's not attached to a line in the pull request diff
	// see https://docs.github.com/en/rest/reference/pulls#create-a-review-comment-for-a-pull-request
	_, _, err = a.github.Issues.CreateComment(ctx, org, repo, prEvent.PullRequest.Number, &github.IssueComment{
		Body: github.String(comment),
	})

	if err != nil {
		return fmt.Errorf("error decorating pull request: %s", err)
	}

	return nil
}

func statusMessage(pass bool) string {
	if pass {
		return "✅ (PASSED)"
	}

	return "❌ (FAILED)"
}
