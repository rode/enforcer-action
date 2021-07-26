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
	"errors"
	"fmt"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rode/enforcer-action/config"
	rode "github.com/rode/rode/proto/v1alpha1"
	"github.com/rode/rode/proto/v1alpha1fakes"
	"google.golang.org/grpc"
)

var _ = Describe("EnforcerAction", func() {
	var (
		ctx                 = context.Background()
		conf                *config.Config
		rodeClient          *v1alpha1fakes.FakeRodeClient
		action              *EnforcerAction
		expectedPolicyGroup string
		expectedResourceUri string
	)

	BeforeEach(func() {
		rodeClient = &v1alpha1fakes.FakeRodeClient{}
		expectedPolicyGroup = fake.LetterN(10)
		expectedResourceUri = fake.URL()
		conf = &config.Config{
			Enforce:     true,
			ResourceUri: expectedResourceUri,
			PolicyGroup: expectedPolicyGroup,
			GitHub: &config.GitHubConfig{
				ServerUrl:  fake.URL(),
				Repository: fmt.Sprintf("%s/%s", fake.LetterN(10), fake.LetterN(10)),
				RunId:      fake.Number(10, 100),
			},
		}

		action = NewEnforcerAction(logger, conf, rodeClient)
	})

	Describe("Run", func() {
		var (
			actualResult           *ActionResult
			actualError            error
			policyEvaluationsCount int

			resourceEvaluationResult *rode.ResourceEvaluationResult
			resourceEvaluationError  error
			expectedPolicyNames      map[string]string
		)

		BeforeEach(func() {
			policyEvaluationsCount = fake.Number(2, 5)
			resourceEvaluationResult = &rode.ResourceEvaluationResult{
				ResourceEvaluation: &rode.ResourceEvaluation{
					Id:   fake.UUID(),
					Pass: true,
				},
			}
			resourceEvaluationError = nil
			expectedPolicyNames = map[string]string{}

			for i := 0; i < policyEvaluationsCount; i++ {
				evaluation := &rode.PolicyEvaluation{
					Id:                   fake.UUID(),
					ResourceEvaluationId: resourceEvaluationResult.ResourceEvaluation.Id,
					Pass:                 fake.Bool(),
					PolicyVersionId:      fake.UUID(),
					Violations: []*rode.EvaluatePolicyViolation{
						{
							Message: fake.Word(),
						},
						{
							Message: fake.Word(),
						},
					},
				}
				expectedPolicyNames[evaluation.Id] = fake.Word()
				resourceEvaluationResult.PolicyEvaluations = append(resourceEvaluationResult.PolicyEvaluations, evaluation)
			}

			rodeClient.GetPolicyStub = func(_ context.Context, request *rode.GetPolicyRequest, _ ...grpc.CallOption) (*rode.Policy, error) {
				policyName := expectedPolicyNames[request.Id]

				return &rode.Policy{
					Id:   request.Id,
					Name: policyName,
				}, nil
			}
		})

		JustBeforeEach(func() {
			rodeClient.EvaluateResourceReturns(resourceEvaluationResult, resourceEvaluationError)

			actualResult, actualError = action.Run(ctx)
		})

		When("the resource evaluation passes", func() {
			It("should call Rode to evaluate the resource", func() {
				expectedSourceUrl := strings.Join([]string{
					conf.GitHub.ServerUrl,
					conf.GitHub.Repository,
					"actions",
					"runs",
					strconv.Itoa(conf.GitHub.RunId),
				}, "/")

				Expect(rodeClient.EvaluateResourceCallCount()).To(Equal(1))

				_, actualRequest, _ := rodeClient.EvaluateResourceArgsForCall(0)
				Expect(actualRequest.ResourceUri).To(Equal(expectedResourceUri))
				Expect(actualRequest.PolicyGroup).To(Equal(expectedPolicyGroup))
				Expect(actualRequest.Source.Name).To(Equal("enforcer-action"))
				Expect(actualRequest.Source.Url).To(Equal(expectedSourceUrl))
			})

			It("should return the result of the evaluation", func() {
				Expect(actualResult.Pass).To(BeTrue())
				Expect(actualResult.FailBuild).To(BeFalse())
				Expect(actualResult.EvaluationReport).To(ContainSubstring("Resource PASSED evaluation"))
			})

			It("should fetch the policy names to include in the output", func() {
				Expect(rodeClient.GetPolicyCallCount()).To(Equal(policyEvaluationsCount))

				for i := 0; i < policyEvaluationsCount; i++ {
					expectedPolicyId := resourceEvaluationResult.PolicyEvaluations[i].PolicyVersionId
					_, actualRequest, _ := rodeClient.GetPolicyArgsForCall(i)

					Expect(actualRequest.Id).To(Equal(expectedPolicyId))
					Expect(actualResult.EvaluationReport).To(ContainSubstring(expectedPolicyNames[expectedPolicyId]))
				}
			})

			It("should not return an error", func() {
				Expect(actualError).To(BeNil())
			})
		})

		When("the resource fails evaluation", func() {
			BeforeEach(func() {
				resourceEvaluationResult.ResourceEvaluation.Pass = false
			})

			It("should fail the build", func() {
				Expect(actualResult.Pass).To(BeFalse())
				Expect(actualResult.FailBuild).To(BeTrue())
				Expect(actualResult.EvaluationReport).To(ContainSubstring("Resource FAILED evaluation"))
			})

			When("enforcement is disabled", func() {
				BeforeEach(func() {
					conf.Enforce = false
				})

				It("should not fail the build", func() {
					Expect(actualResult.Pass).To(BeFalse())
					Expect(actualResult.FailBuild).To(BeFalse())
				})

				It("should not return an error", func() {
					Expect(actualError).NotTo(HaveOccurred())
				})
			})
		})

		When("an error occurs evaluating the resource", func() {
			BeforeEach(func() {
				resourceEvaluationError = errors.New(fake.Word())
			})

			It("should return an error", func() {
				Expect(actualResult).To(BeNil())
				Expect(actualError).To(MatchError(ContainSubstring("error evaluating resource")))
			})
		})

		When("an error occurs fetching the policy", func() {
			BeforeEach(func() {
				rodeClient.GetPolicyReturns(nil, errors.New("get policy error"))
			})

			It("should return an error", func() {
				Expect(actualResult).To(BeNil())
				Expect(actualError).To(HaveOccurred())
			})
		})
	})
})
