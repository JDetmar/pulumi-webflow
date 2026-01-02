//go:build all || typescript || python || go || csharp || java
// +build all typescript python go csharp java

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptRedirectExample tests the TypeScript Redirect example
func TestTypeScriptRedirectExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("redirect", "typescript"),
		opttest.YarnLink("pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
}

// TestPythonRedirectExample tests the Python Redirect example
func TestPythonRedirectExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("redirect", "python"),
		opttest.PythonLink("../sdk/python"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployed_site_id"].Value == nil {
		t.Error("Expected deployed_site_id output")
	}
}

// TestGoRedirectExample tests the Go Redirect example
func TestGoRedirectExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("redirect", "go"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
}

// TestCSharpRedirectExample tests the C# Redirect example
func TestCSharpRedirectExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("redirect", "csharp"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
}

// TestJavaRedirectExample tests the Java Redirect example
func TestJavaRedirectExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("redirect", "java"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
}
