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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rode/evaluate-policy-action/config"
	"github.com/rode/evaluate-policy-action/mocks"
	rode "github.com/rode/rode/proto/v1alpha1"
)

var _ = Describe("EvaluatePolicyAction", func() {
	var (
		ctx        = context.Background()
		conf       *config.Config
		rodeClient *mocks.FakeRodeClient
		action     *EvaluatePolicyAction
	)

	BeforeEach(func() {
		rodeClient = &mocks.FakeRodeClient{}
		conf = &config.Config{
			PolicyId:    fake.UUID(),
			ResourceUri: fake.URL(),
		}

		action = NewEvaluatePolicyAction(logger, conf, rodeClient)
	})

	Describe("Run", func() {
		var (
			actualPass  bool
			actualError error
		)

		JustBeforeEach(func() {
			actualPass, actualError = action.Run(ctx)
		})

		When("policy is successfully evaluated", func() {
			var expectedPass bool

			BeforeEach(func() {
				expectedPass = fake.Bool()
				rodeClient.EvaluatePolicyReturns(&rode.EvaluatePolicyResponse{Pass: expectedPass}, nil)
			})

			It("should call Rode to evaluate policy", func() {
				Expect(rodeClient.EvaluatePolicyCallCount()).To(Equal(1))

				_, actualRequest, _ := rodeClient.EvaluatePolicyArgsForCall(0)
				Expect(actualRequest.ResourceUri).To(Equal(conf.ResourceUri))
				Expect(actualRequest.Policy).To(Equal(conf.PolicyId))
			})

			It("should return the result of the evaluation", func() {
				Expect(actualPass).To(Equal(expectedPass))
			})

			It("should not return an error", func() {
				Expect(actualError).To(BeNil())
			})
		})

		Context("policyName is set", func() {
			var (
				expectedPolicyName   string
				expectedPolicyId     string
				listPoliciesError    error
				listPoliciesResponse *rode.ListPoliciesResponse
			)

			BeforeEach(func() {
				expectedPolicyId = fake.UUID()
				expectedPolicyName = fake.Word()

				conf.PolicyName = expectedPolicyName
				conf.PolicyId = ""

				listPoliciesError = nil
				listPoliciesResponse = &rode.ListPoliciesResponse{
					Policies: []*rode.Policy{
						{
							Id: expectedPolicyId,
							Policy: &rode.PolicyEntity{
								Name: expectedPolicyName,
							},
						},
					},
				}
			})

			When("when there's a single matching policy", func() {
				BeforeEach(func() {
					rodeClient.ListPoliciesReturns(listPoliciesResponse, listPoliciesError)
					rodeClient.EvaluatePolicyReturns(&rode.EvaluatePolicyResponse{Pass: true}, nil)
				})

				It("should use Rode to search policies by name", func() {
					Expect(rodeClient.ListPoliciesCallCount()).To(Equal(1))
					_, actualRequest, _ := rodeClient.ListPoliciesArgsForCall(0)

					Expect(actualRequest.Filter).To(Equal(`name == "` + expectedPolicyName + `"`))
				})

				It("should use the policy id during evaluation", func() {
					Expect(rodeClient.EvaluatePolicyCallCount()).To(Equal(1))

					_, actualRequest, _ := rodeClient.EvaluatePolicyArgsForCall(0)

					Expect(actualRequest.Policy).To(Equal(expectedPolicyId))
				})
			})

			When("no policies match the name", func() {
				BeforeEach(func() {
					listPoliciesResponse.Policies = []*rode.Policy{}
					rodeClient.ListPoliciesReturns(listPoliciesResponse, listPoliciesError)
				})

				It("should not pass", func() {
					Expect(actualPass).To(BeFalse())
				})

				It("should return an error", func() {
					Expect(actualError).To(HaveOccurred())
					Expect(actualError.Error()).To(ContainSubstring("unable to find a policy"))
				})
			})

			When("multiple policies match the name", func() {
				BeforeEach(func() {
					listPoliciesResponse.Policies = append(listPoliciesResponse.Policies, &rode.Policy{Id: fake.UUID()})
					rodeClient.ListPoliciesReturns(listPoliciesResponse, listPoliciesError)
					rodeClient.EvaluatePolicyReturns(&rode.EvaluatePolicyResponse{Pass: true}, nil)
				})

				It("should use the first policy in the response", func() {
					Expect(rodeClient.EvaluatePolicyCallCount()).To(Equal(1))

					_, actualRequest, _ := rodeClient.EvaluatePolicyArgsForCall(0)

					Expect(actualRequest.Policy).To(Equal(expectedPolicyId))
				})
			})

			When("an error occurs searching policies", func() {
				BeforeEach(func() {
					listPoliciesResponse = nil
					listPoliciesError = errors.New(fake.Word())

					rodeClient.ListPoliciesReturns(listPoliciesResponse, listPoliciesError)
				})

				It("should not pass", func() {
					Expect(actualPass).To(BeFalse())
				})

				It("should return an error", func() {
					Expect(actualError).To(HaveOccurred())
					Expect(actualError.Error()).To(ContainSubstring("error fetching policies"))
				})
			})
		})

		When("an error occurs evaluating policy", func() {
			BeforeEach(func() {
				rodeClient.EvaluatePolicyReturns(nil, errors.New(fake.Word()))
			})

			It("should not pass", func() {
				Expect(actualPass).To(BeFalse())
			})

			It("should return an error", func() {
				Expect(actualError).To(HaveOccurred())
				Expect(actualError.Error()).To(ContainSubstring("error evaluating policy"))
			})
		})
	})
})
