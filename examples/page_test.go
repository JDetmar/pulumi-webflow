//go:build all || typescript || python || go || csharp || java
// +build all typescript python go csharp java

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptPageExample tests the TypeScript Page example
func TestTypeScriptPageExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("page", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	test.Preview(t)
	result := test.Up(t)

	// Verify we got page data
	if result.Outputs["pageCount"].Value == nil {
		t.Error("Expected pageCount output")
	}
	if result.Outputs["pageIds"].Value == nil {
		t.Error("Expected pageIds output")
	}
	if result.Outputs["sitePages"].Value == nil {
		t.Error("Expected sitePages output")
	}
}

// TestPythonPageExample tests the Python Page example
func TestPythonPageExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("page", "python"),
		opttest.PythonLink("../sdk/python"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	// Verify we got page data
	if result.Outputs["page_count"].Value == nil {
		t.Error("Expected page_count output")
	}
	if result.Outputs["page_ids"].Value == nil {
		t.Error("Expected page_ids output")
	}
	if result.Outputs["site_pages"].Value == nil {
		t.Error("Expected site_pages output")
	}
}

// TestGoPageExample tests the Go Page example
func TestGoPageExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("page", "go"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	// Verify we got page data
	if result.Outputs["pageCount"].Value == nil {
		t.Error("Expected pageCount output")
	}
	if result.Outputs["pageIds"].Value == nil {
		t.Error("Expected pageIds output")
	}
	if result.Outputs["sitePages"].Value == nil {
		t.Error("Expected sitePages output")
	}
}
