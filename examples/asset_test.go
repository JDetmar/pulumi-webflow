//go:build all || typescript || python || go || csharp || java
// +build all typescript python go csharp java

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptAssetExample tests the TypeScript Asset example
func TestTypeScriptAssetExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("asset", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["logoAssetId"].Value == nil {
		t.Error("Expected logoAssetId output")
	}
}

// TestPythonAssetExample tests the Python Asset example
func TestPythonAssetExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("asset", "python"),
		opttest.PythonLink("../sdk/python"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["logo_asset_id"].Value == nil {
		t.Error("Expected logo_asset_id output")
	}
}

// TestGoAssetExample tests the Go Asset example
func TestGoAssetExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("asset", "go"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["logoAssetId"].Value == nil {
		t.Error("Expected logoAssetId output")
	}
}

// TestCSharpAssetExample tests the C# Asset example
func TestCSharpAssetExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("asset", "csharp"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["logoAssetId"].Value == nil {
		t.Error("Expected logoAssetId output")
	}
}

// TestJavaAssetExample tests the Java Asset example
func TestJavaAssetExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("asset", "java"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["logoAssetId"].Value == nil {
		t.Error("Expected logoAssetId output")
	}
}
