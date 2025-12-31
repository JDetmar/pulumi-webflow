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
		opttest.YarnLink("pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)
	defer test.Cleanup(t)

	// Preview
	test.Preview(t)

	// Deploy
	test.Up(t)

	// Verify outputs
	outputs := test.GetStackOutputs(t)
	if outputs["deployedSiteId"] == nil {
		t.Error("Expected deployedSiteId output")
	}

	// Cleanup
	test.Destroy(t)
}

// TestPythonRobotsTxtExample tests the Python RobotsTxt example
func TestPythonRobotsTxtExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("robotstxt", "python"),
		opttest.Pip("pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)
	defer test.Cleanup(t)

	// Preview
	test.Preview(t)

	// Deploy
	test.Up(t)

	// Verify outputs
	outputs := test.GetStackOutputs(t)
	if outputs["deployed_site_id"] == nil {
		t.Error("Expected deployed_site_id output")
	}

	// Cleanup
	test.Destroy(t)
}

// TestGoRobotsTxtExample tests the Go RobotsTxt example
func TestGoRobotsTxtExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("robotstxt", "go"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)
	defer test.Cleanup(t)

	// Preview
	test.Preview(t)

	// Deploy
	test.Up(t)

	// Verify outputs
	outputs := test.GetStackOutputs(t)
	if outputs["deployedSiteId"] == nil {
		t.Error("Expected deployedSiteId output")
	}

	// Cleanup
	test.Destroy(t)
}

// TestCSharpRobotsTxtExample tests the C# RobotsTxt example
func TestCSharpRobotsTxtExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("robotstxt", "csharp"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)
	defer test.Cleanup(t)

	// Preview
	test.Preview(t)

	// Deploy
	test.Up(t)

	// Verify outputs
	outputs := test.GetStackOutputs(t)
	if outputs["deployedSiteId"] == nil {
		t.Error("Expected deployedSiteId output")
	}

	// Cleanup
	test.Destroy(t)
}

// TestJavaRobotsTxtExample tests the Java RobotsTxt example
func TestJavaRobotsTxtExample(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join("robotstxt", "java"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)
	defer test.Cleanup(t)

	// Preview
	test.Preview(t)

	// Deploy
	test.Up(t)

	// Verify outputs
	outputs := test.GetStackOutputs(t)
	if outputs["deployedSiteId"] == nil {
		t.Error("Expected deployedSiteId output")
	}

	// Cleanup
	test.Destroy(t)
}
