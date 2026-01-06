//go:build all || typescript
// +build all typescript

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptSiteCustomCodeExample tests the TypeScript SiteCustomCode example
func TestTypeScriptSiteCustomCodeExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("sitecustomcode", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
	if result.Outputs["siteScriptsCreatedOn"].Value == nil {
		t.Error("Expected siteScriptsCreatedOn output")
	}
	if result.Outputs["appliedScriptCount"].Value == nil {
		t.Error("Expected appliedScriptCount output")
	}
}
