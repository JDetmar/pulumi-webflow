//go:build all || typescript || python || go || csharp || java
// +build all typescript python go csharp java

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptRobotsTxtExample tests the TypeScript RobotsTxt example
func TestTypeScriptRobotsTxtExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("robotstxt", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	// Preview
	test.Preview(t)

	// Deploy


	// Verify outputs
	result := test.Up(t)
	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}

	// Cleanup
}

// TestPythonRobotsTxtExample tests the Python RobotsTxt example
func TestPythonRobotsTxtExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("robotstxt", "python"),
		opttest.PythonLink("../sdk/python"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	// Preview
	test.Preview(t)

	// Deploy


	// Verify outputs
	result := test.Up(t)
	if result.Outputs["deployed_site_id"].Value == nil {
		t.Error("Expected deployed_site_id output")
	}

	// Cleanup
}

// TestGoRobotsTxtExample tests the Go RobotsTxt example
func TestGoRobotsTxtExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("robotstxt", "go"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	// Preview
	test.Preview(t)

	// Deploy


	// Verify outputs
	result := test.Up(t)
	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}

	// Cleanup
}

// TestCSharpRobotsTxtExample tests the C# RobotsTxt example
func TestCSharpRobotsTxtExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("robotstxt", "csharp"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	// Preview
	test.Preview(t)

	// Deploy


	// Verify outputs
	result := test.Up(t)
	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}

	// Cleanup
}

// TestJavaRobotsTxtExample tests the Java RobotsTxt example
func TestJavaRobotsTxtExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("robotstxt", "java"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	// Preview
	test.Preview(t)

	// Deploy


	// Verify outputs
	result := test.Up(t)
	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}

	// Cleanup
}
