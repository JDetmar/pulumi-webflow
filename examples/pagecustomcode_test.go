//go:build all || typescript
// +build all typescript

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptPageCustomCodeExample tests the TypeScript PageCustomCode example
func TestTypeScriptPageCustomCodeExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("pagecustomcode", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
	if result.Outputs["landingPageScriptsCreatedOn"].Value == nil {
		t.Error("Expected landingPageScriptsCreatedOn output")
	}
	if result.Outputs["conversionTrackingScriptId"].Value == nil {
		t.Error("Expected conversionTrackingScriptId output")
	}
}
