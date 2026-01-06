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

	test.Preview(t)
	result := test.Up(t)

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
}
