// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package tests

import (
	"context"
	"testing"

	xyz "github.com/JDetmar/pulumi-webflow/provider"
	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

func TestSitePreviewCreate(t *testing.T) {
	t.Parallel()

	prov := provider(t)

	response, err := prov.Create(p.CreateRequest{
		Urn: urn("Site"),
		Properties: property.NewMap(map[string]property.Value{
			"workspaceId": property.New("5f0c8c9e1c9d440000e8d8c3"),
			"displayName": property.New("My Pulumi Site"),
			"shortName":   property.New("my-pulumi-site"),
			"timeZone":    property.New("UTC"),
		}),
		DryRun: true, // Avoid network calls; preview path exercises validation and state wiring
	})

	require.NoError(t, err)
	assert.Len(t, response.ID, 24)
	assert.Equal(t, "My Pulumi Site", response.Properties.Get("displayName").AsString())
}

// urn is a helper function to build an urn for running integration tests.
func urn(typ string) resource.URN {
	return resource.NewURN(
		"stack",
		"proj",
		"",
		tokens.Type(xyz.Name+":index:"+typ),
		"name",
	)
}

// Create a test server.
func provider(t *testing.T) integration.Server {
	s, err := integration.NewServer(
		context.Background(),
		xyz.Name,
		semver.MustParse("1.0.0"),
		integration.WithProvider(xyz.Provider()),
	)
	require.NoError(t, err)
	return s
}
