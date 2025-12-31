//go:build all || typescript || python || go || csharp || java
// +build all typescript python go csharp java

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptSiteExample tests the TypeScript Site example
func TestTypeScriptSiteExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("site", "typescript"),
		opttest.YarnLink("pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)
	defer test.Cleanup(t)

	// Set required config
	test.SetConfig(t, "displayName", "Test Site")
	test.SetConfig(t, "shortName", "test-site")

	test.Preview(t)
	test.Up(t)

	outputs := test.GetStackOutputs(t)
	if outputs["basicSiteId"] == nil {
		t.Error("Expected basicSiteId output")
	}

	test.Destroy(t)
}

// TestPythonSiteExample tests the Python Site example
func TestPythonSiteExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("site", "python"),
		opttest.Pip("pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)
	defer test.Cleanup(t)

	// Set required config
	test.SetConfig(t, "displayName", "Test Site")
	test.SetConfig(t, "shortName", "test-site")

	test.Preview(t)
	test.Up(t)

	outputs := test.GetStackOutputs(t)
	if outputs["basic_site_id"] == nil {
		t.Error("Expected basic_site_id output")
	}

	test.Destroy(t)
}

// TestGoSiteExample tests the Go Site example
func TestGoSiteExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("site", "go"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)
	defer test.Cleanup(t)

	// Set required config
	test.SetConfig(t, "displayName", "Test Site")
	test.SetConfig(t, "shortName", "test-site")

	test.Preview(t)
	test.Up(t)

	outputs := test.GetStackOutputs(t)
	if outputs["basicSiteId"] == nil {
		t.Error("Expected basicSiteId output")
	}

	test.Destroy(t)
}

// TestCSharpSiteExample tests the C# Site example
func TestCSharpSiteExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("site", "csharp"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)
	defer test.Cleanup(t)

	// Set required config
	test.SetConfig(t, "displayName", "Test Site")
	test.SetConfig(t, "shortName", "test-site")

	test.Preview(t)
	test.Up(t)

	outputs := test.GetStackOutputs(t)
	if outputs["basicSiteId"] == nil {
		t.Error("Expected basicSiteId output")
	}

	test.Destroy(t)
}

// TestJavaSiteExample tests the Java Site example
func TestJavaSiteExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("site", "java"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)
	defer test.Cleanup(t)

	// Set required config
	test.SetConfig(t, "displayName", "Test Site")
	test.SetConfig(t, "shortName", "test-site")

	test.Preview(t)
	test.Up(t)

	outputs := test.GetStackOutputs(t)
	if outputs["basicSiteId"] == nil {
		t.Error("Expected basicSiteId output")
	}

	test.Destroy(t)
}
