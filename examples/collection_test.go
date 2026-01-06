//go:build all || typescript || python || go || csharp || java
// +build all typescript python go csharp java

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptCollectionExample tests the TypeScript Collection example
func TestTypeScriptCollectionExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("collection", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
	if result.Outputs["blogCollectionId"].Value == nil {
		t.Error("Expected blogCollectionId output")
	}
}

// TestPythonCollectionExample tests the Python Collection example
func TestPythonCollectionExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("collection", "python"),
		opttest.PythonLink("../sdk/python"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployed_site_id"].Value == nil {
		t.Error("Expected deployed_site_id output")
	}
	if result.Outputs["blog_collection_id"].Value == nil {
		t.Error("Expected blog_collection_id output")
	}
}

// TestGoCollectionExample tests the Go Collection example
func TestGoCollectionExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("collection", "go"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
	if result.Outputs["blogCollectionId"].Value == nil {
		t.Error("Expected blogCollectionId output")
	}
}

// TestCSharpCollectionExample tests the C# Collection example
func TestCSharpCollectionExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("collection", "csharp"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
	if result.Outputs["blogCollectionId"].Value == nil {
		t.Error("Expected blogCollectionId output")
	}
}

// TestJavaCollectionExample tests the Java Collection example
func TestJavaCollectionExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("collection", "java"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
	if result.Outputs["blogCollectionId"].Value == nil {
		t.Error("Expected blogCollectionId output")
	}
}
