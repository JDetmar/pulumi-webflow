//go:build all || typescript
// +build all typescript

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptCollectionFieldExample tests the TypeScript CollectionField example
func TestTypeScriptCollectionFieldExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("collectionfield", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedCollectionId"].Value == nil {
		t.Error("Expected deployedCollectionId output")
	}
	if result.Outputs["titleFieldId"].Value == nil {
		t.Error("Expected titleFieldId output")
	}
	if result.Outputs["contentFieldId"].Value == nil {
		t.Error("Expected contentFieldId output")
	}
	if result.Outputs["summary"].Value == nil {
		t.Error("Expected summary output")
	}
}
