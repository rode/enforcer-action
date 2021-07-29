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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v35/github"
	"github.com/rode/enforcer-action/action"
	"github.com/rode/enforcer-action/config"
	"github.com/rode/rode/common"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

func newLogger() (*zap.Logger, error) {
	c := zap.NewDevelopmentConfig()
	c.DisableCaller = true
	c.EncoderConfig.LevelKey = ""
	c.EncoderConfig.TimeKey = ""

	return c.Build()
}

func setOutputVariable(name string, value interface{}) {
	fmt.Printf("\n::set-output name=%s::%v\n", name, value)
}

func fatal(message string) {
	fmt.Println(message)
	os.Exit(1)
}

type staticCredential struct {
	token                    string
	requireTransportSecurity bool
}

func (s *staticCredential) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + s.token,
	}, nil
}

func (s *staticCredential) RequireTransportSecurity() bool {
	return s.requireTransportSecurity
}

func newGitHubClient(c *config.GitHubConfig) *github.Client {
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: c.Token,
		},
	)

	return github.NewClient(oauth2.NewClient(context.Background(), tokenSource))
}

func writeEvaluationReport(logger *zap.Logger, report string) string {
	reportDir := os.TempDir()
	file, err := os.CreateTemp(reportDir, "report-*.json")
	if err != nil {
		logger.Fatal("error opening temporary file", zap.Error(err))
	}
	defer file.Close()

	if _, err = file.WriteString(report); err != nil {
		logger.Fatal("error writing report to disk", zap.Error(err))
	}

	return file.Name()
}

func main() {
	ctx := context.Background()
	c, err := config.Build(os.Args[0], os.Args[1:])
	if err != nil {
		fatal(fmt.Sprintf("unable to build config: %s", err))
	}

	logger, err := newLogger()
	if err != nil {
		fatal(fmt.Sprintf("failed to create logger: %s", err))
	}

	var dialOptions []grpc.DialOption
	if c.AccessToken != "" {
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(&staticCredential{
			token:                    c.AccessToken,
			requireTransportSecurity: !c.ClientConfig.Rode.DisableTransportSecurity,
		}))
	}

	rodeClient, err := common.NewRodeClient(c.ClientConfig, dialOptions...)
	if err != nil {
		logger.Fatal("failed to create rode client", zap.Error(err))
	}

	enforcer := action.NewEnforcerAction(logger, c, rodeClient, newGitHubClient(c.GitHub))
	result, err := enforcer.Run(ctx)
	if err != nil {
		logger.Fatal("error evaluating resource", zap.Error(err))
	}

	logger.Info(result.EvaluationReport)
	reportPath := writeEvaluationReport(logger, result.EvaluationReport)

	logger.Info("Wrote evaluation report", zap.String("report", reportPath))

	setOutputVariable("pass", result.Pass)
	setOutputVariable("reportPath", reportPath)

	if result.FailBuild {
		os.Exit(1)
	}
}
