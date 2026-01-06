//go:build all || typescript
// +build all typescript

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptPageContentExample tests the TypeScript PageContent example
func TestTypeScriptPageContentExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("pagecontent", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedPageId"].Value == nil {
		t.Error("Expected deployedPageId output")
	}

	if result.Outputs["heroContentId"].Value == nil {
		t.Error("Expected heroContentId output")
	}

	if result.Outputs["heroLastUpdated"].Value == nil {
		t.Error("Expected heroLastUpdated output")
	}
}
