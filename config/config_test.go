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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/rode/rode/common"
)

var _ = Describe("Config", func() {
	Context("Build", func() {
		var (
			expectedPolicyGroup = fake.URL()
			expectedResourceUri = fake.LetterN(10)
		)

		type testCase struct {
			flags       []string
			expectError bool
			expected    *Config
		}

		DescribeTable("configuration", func(tc *testCase) {
			c, err := Build("evaluate-resource-action", tc.flags)

			if tc.expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(c).To(Equal(tc.expected))
			}
		},
			Entry("minimum configuration", &testCase{
				flags: []string{
					"--policy-group=" + expectedPolicyGroup,
					"--resource-uri=" + expectedResourceUri,
				},
				expected: &Config{
					Enforce: true,
					GitHub:  &GitHubConfig{},
					ClientConfig: &common.ClientConfig{
						Rode: &common.RodeClientConfig{
							Host: "rode:50051",
						},
						OIDCAuth:  &common.OIDCAuthConfig{},
						BasicAuth: &common.BasicAuthConfig{},
					},
					ResourceUri: expectedResourceUri,
					PolicyGroup: expectedPolicyGroup,
				},
			}),

			Entry("whitespace in policy group name", &testCase{
				flags: []string{
					"--policy-group=" + "   " + expectedPolicyGroup + "   ",
					"--resource-uri=" + expectedResourceUri,
				},
				expected: &Config{
					Enforce: true,
					GitHub:  &GitHubConfig{},
					ClientConfig: &common.ClientConfig{
						Rode: &common.RodeClientConfig{
							Host: "rode:50051",
						},
						OIDCAuth:  &common.OIDCAuthConfig{},
						BasicAuth: &common.BasicAuthConfig{},
					},
					ResourceUri: expectedResourceUri,
					PolicyGroup: expectedPolicyGroup,
				},
			}),
			Entry("missing policy group", &testCase{
				flags: []string{
					"--resource-uri=" + expectedResourceUri,
				},
				expectError: true,
			}),

			Entry("missing resource uri", &testCase{
				flags: []string{
					"--policy-group=" + expectedPolicyGroup,
				},
				expectError: true,
			}),
			Entry("invalid flag value", &testCase{
				flags:       []string{"--enforce=foo"},
				expectError: true,
			}),
		)
	})
})
