//go:build all || typescript
// +build all typescript

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptAssetFolderExample tests the TypeScript AssetFolder example
func TestTypeScriptAssetFolderExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("assetfolder", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}

	if result.Outputs["imagesFolderId"].Value == nil {
		t.Error("Expected imagesFolderId output")
	}

	if result.Outputs["documentsFolderId"].Value == nil {
		t.Error("Expected documentsFolderId output")
	}

	if result.Outputs["iconsFolderId"].Value == nil {
		t.Error("Expected iconsFolderId output")
	}

	if result.Outputs["heroImagesFolderId"].Value == nil {
		t.Error("Expected heroImagesFolderId output (nested folder)")
	}

	if result.Outputs["bulkFolderIds"].Value == nil {
		t.Error("Expected bulkFolderIds output")
	}
}
