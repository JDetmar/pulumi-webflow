//go:build all || typescript
// +build all typescript

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptUserExample tests the TypeScript User example
func TestTypeScriptUserExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("user", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	// Set required config - siteId is required by the User resource
	test.SetConfig(t, "siteId", "580e63e98c9a982ac9b8b741")

	test.Preview(t)
	result := test.Up(t)

	// Validate basic user outputs
	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
	if result.Outputs["basicUserId"].Value == nil {
		t.Error("Expected basicUserId output")
	}
	if result.Outputs["basicUserEmail"].Value == nil {
		t.Error("Expected basicUserEmail output")
	}
	if result.Outputs["basicUserStatus"].Value == nil {
		t.Error("Expected basicUserStatus output")
	}
	if result.Outputs["basicUserVerified"].Value == nil {
		t.Error("Expected basicUserVerified output")
	}

	// Validate named user outputs
	if result.Outputs["namedUserId"].Value == nil {
		t.Error("Expected namedUserId output")
	}
	if result.Outputs["namedUserName"].Value == nil {
		t.Error("Expected namedUserName output")
	}

	// Validate premium user outputs
	if result.Outputs["premiumUserId"].Value == nil {
		t.Error("Expected premiumUserId output")
	}
	if result.Outputs["premiumUserGroups"].Value == nil {
		t.Error("Expected premiumUserGroups output")
	}
	if result.Outputs["premiumUserCreated"].Value == nil {
		t.Error("Expected premiumUserCreated output")
	}

	// Validate beta tester outputs
	if result.Outputs["betaTesterUserId"].Value == nil {
		t.Error("Expected betaTesterUserId output")
	}
	if result.Outputs["betaTesterStatus"].Value == nil {
		t.Error("Expected betaTesterStatus output")
	}

	// Validate power user outputs
	if result.Outputs["powerUserId"].Value == nil {
		t.Error("Expected powerUserId output")
	}
	if result.Outputs["powerUserGroups"].Value == nil {
		t.Error("Expected powerUserGroups output")
	}
}
