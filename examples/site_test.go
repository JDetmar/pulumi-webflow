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
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	// Set required config
	test.SetConfig(t, "displayName", "Test Site")
	test.SetConfig(t, "shortName", "test-site")

	test.Preview(t)


	result := test.Up(t)
	if result.Outputs["basicSiteId"].Value == nil {
		t.Error("Expected basicSiteId output")
	}

}

// TestPythonSiteExample tests the Python Site example
func TestPythonSiteExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("site", "python"),
		opttest.PythonLink("../sdk/python"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	// Set required config
	test.SetConfig(t, "displayName", "Test Site")
	test.SetConfig(t, "shortName", "test-site")

	test.Preview(t)


	result := test.Up(t)
	if result.Outputs["basic_site_id"].Value == nil {
		t.Error("Expected basic_site_id output")
	}

}

// TestGoSiteExample tests the Go Site example
func TestGoSiteExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("site", "go"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	// Set required config
	test.SetConfig(t, "displayName", "Test Site")
	test.SetConfig(t, "shortName", "test-site")

	test.Preview(t)


	result := test.Up(t)
	if result.Outputs["basicSiteId"].Value == nil {
		t.Error("Expected basicSiteId output")
	}

}

// TestCSharpSiteExample tests the C# Site example
func TestCSharpSiteExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("site", "csharp"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	// Set required config
	test.SetConfig(t, "displayName", "Test Site")
	test.SetConfig(t, "shortName", "test-site")

	test.Preview(t)


	result := test.Up(t)
	if result.Outputs["basicSiteId"].Value == nil {
		t.Error("Expected basicSiteId output")
	}

}

// TestJavaSiteExample tests the Java Site example
func TestJavaSiteExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("site", "java"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	// Set required config
	test.SetConfig(t, "displayName", "Test Site")
	test.SetConfig(t, "shortName", "test-site")

	test.Preview(t)


	result := test.Up(t)
	if result.Outputs["basicSiteId"].Value == nil {
		t.Error("Expected basicSiteId output")
	}

}
