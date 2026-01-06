//go:build all || typescript
// +build all typescript

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptRegisteredScriptExample tests the TypeScript RegisteredScript example
func TestTypeScriptRegisteredScriptExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("registeredscript", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
	if result.Outputs["analyticsScriptId"].Value == nil {
		t.Error("Expected analyticsScriptId output")
	}
	if result.Outputs["cmsSliderScriptId"].Value == nil {
		t.Error("Expected cmsSliderScriptId output")
	}
}
