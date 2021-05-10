package config

import (
	"context"
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sethvargo/go-envconfig"
)

var _ = Describe("Config", func() {
	Context("Build", func() {
		var (
			ctx = context.Background()

			configMap            map[string]string
			expectedEnforce      bool
			expectedRodeInsecure bool

			actualConfig *Config
			actualError  error
		)

		BeforeEach(func() {
			expectedEnforce = fake.Bool()
			expectedRodeInsecure = fake.Bool()

			configMap = map[string]string{
				"ENFORCE":       strconv.FormatBool(expectedEnforce),
				"POLICY_ID":     fake.UUID(),
				"POLICY_NAME":   "",
				"RESOURCE_URI":  fake.URL(),
				"RODE_HOST":     fake.DomainName(),
				"RODE_INSECURE": strconv.FormatBool(expectedRodeInsecure),
			}
		})

		JustBeforeEach(func() {
			actualConfig, actualError = Build(ctx, envconfig.MapLookuper(configMap))
		})

		Describe("valid configuration", func() {
			It("should not return an error", func() {
				Expect(actualError).NotTo(HaveOccurred())
			})

			It("should map the environment config to the config struct", func() {
				Expect(actualConfig.PolicyId).To(Equal(configMap["POLICY_ID"]))
				Expect(actualConfig.PolicyName).To(BeEmpty())

				Expect(actualConfig.Enforce).To(Equal(expectedEnforce))
				Expect(actualConfig.ResourceUri).To(Equal(configMap["RESOURCE_URI"]))

				Expect(actualConfig.Rode.Host).To(Equal(configMap["RODE_HOST"]))
				Expect(actualConfig.Rode.Insecure).To(Equal(expectedRodeInsecure))
			})
		})

		Describe("policyId and policyName are set to the empty string", func() {
			BeforeEach(func() {
				configMap["POLICY_ID"] = ""
				configMap["POLICY_NAME"] = ""
			})

			It("should return an error", func() {
				Expect(actualError).To(MatchError("must set either policyName or policyId"))
			})
		})

		Describe("both policyId and policyName are set", func() {
			BeforeEach(func() {
				configMap["POLICY_ID"] = fake.UUID()
				configMap["POLICY_NAME"] = fake.Word()
			})

			It("should return an error", func() {
				Expect(actualError).To(MatchError("only one of policyId or policyName should be specified"))
			})
		})

		Describe("policyName contains additional whitespace", func() {
			var expectedPolicyName string

			BeforeEach(func() {
				expectedPolicyName = fake.Word()

				configMap["POLICY_ID"] = ""
				configMap["POLICY_NAME"] = fmt.Sprintf(" %s  \n\n", expectedPolicyName)
			})

			It("should strip the whitespace", func() {
				Expect(actualConfig.PolicyName).To(Equal(expectedPolicyName))
			})
		})

		Describe("policyId contains additional whitespace", func() {
			var expectedPolicyId string

			BeforeEach(func() {
				expectedPolicyId = fake.UUID()

				configMap["POLICY_ID"] = fmt.Sprintf(" %s  \n\n", expectedPolicyId)
				configMap["POLICY_NAME"] = ""
			})

			It("should strip the whitespace", func() {
				Expect(actualConfig.PolicyId).To(Equal(expectedPolicyId))
			})
		})

		Describe("Rode config is missing", func() {
			BeforeEach(func() {
				delete(configMap, "RODE_HOST")
			})

			It("should return an error", func() {
				Expect(actualError).To(HaveOccurred())
			})
		})
	})
})
