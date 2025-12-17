// Copyright 2025, Justin Detmar.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tests

import (
	"context"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/integration"
	webflow "github.com/jdetmar/pulumi-webflow/provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
)

// urn is a helper function to build an urn for running integration tests.
func urn(typ string) resource.URN {
	return resource.NewURN("stack", "proj", "",
		tokens.Type("webflow:index:"+typ), "name")
}

// Create a test server.
func provider(t *testing.T) integration.Server {
	s, err := integration.NewServer(
		context.Background(),
		webflow.Name,
		semver.MustParse("0.1.0"),
		integration.WithProvider(webflow.Provider()),
	)
	require.NoError(t, err)
	return s
}
