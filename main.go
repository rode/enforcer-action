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
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/rode/evaluate-policy-action/action"
	"github.com/rode/evaluate-policy-action/config"
	rode "github.com/rode/rode/proto/v1alpha1"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func newRodeClient(logger *zap.Logger, conf *config.Config) (*grpc.ClientConn, rode.RodeClient) {
	dialOptions := []grpc.DialOption{
		grpc.WithBlock(),
	}
	if conf.Rode.Insecure {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	conn, err := grpc.DialContext(ctx, conf.Rode.Host, dialOptions...)
	if err != nil {
		logger.Fatal("failed to establish grpc connection to Rode", zap.Error(err))
	}

	return conn, rode.NewRodeClient(conn)
}

func newLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

func setOutputVariable(name string, value bool) {
	fmt.Printf("::set-output name=%s::%t", name, value)
}

func fatal(message string) {
	fmt.Println(message)
	os.Exit(1)
}

func main() {
	ctx := context.Background()
	c, err := config.Build(ctx, envconfig.OsLookuper())
	if err != nil {
		fatal(fmt.Sprintf("unable to build config: %s", err))
	}

	logger, err := newLogger()
	if err != nil {
		fatal(fmt.Sprintf("failed to create logger: %s", err))
	}

	conn, client := newRodeClient(logger, c)
	defer conn.Close()

	evaluatePolicyAction := action.NewEvaluatePolicyAction(logger, c, client)
	pass, err := evaluatePolicyAction.Run(ctx)
	if err != nil {
		fatal(err.Error())
	}

	setOutputVariable("pass", pass)

	if !pass && c.Enforce {
		logger.Info("Policy evaluation failed")
		os.Exit(1)
	}
}
