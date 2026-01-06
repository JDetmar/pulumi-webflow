//go:build all || typescript || python || go || csharp || java
// +build all typescript python go csharp java

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptCollectionItemExample tests the TypeScript CollectionItem example
func TestTypeScriptCollectionItemExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("collectionitem", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedCollectionId"].Value == nil {
		t.Error("Expected deployedCollectionId output")
	}
	if result.Outputs["draftPostId"].Value == nil {
		t.Error("Expected draftPostId output")
	}
	if result.Outputs["publishedProductId"].Value == nil {
		t.Error("Expected publishedProductId output")
	}
}

// TestPythonCollectionItemExample tests the Python CollectionItem example
func TestPythonCollectionItemExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("collectionitem", "python"),
		opttest.PythonLink("../sdk/python"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployed_collection_id"].Value == nil {
		t.Error("Expected deployed_collection_id output")
	}
	if result.Outputs["draft_post_id"].Value == nil {
		t.Error("Expected draft_post_id output")
	}
	if result.Outputs["published_product_id"].Value == nil {
		t.Error("Expected published_product_id output")
	}
}

// TestGoCollectionItemExample tests the Go CollectionItem example
func TestGoCollectionItemExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("collectionitem", "go"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedCollectionId"].Value == nil {
		t.Error("Expected deployedCollectionId output")
	}
	if result.Outputs["draftPostId"].Value == nil {
		t.Error("Expected draftPostId output")
	}
	if result.Outputs["publishedProductId"].Value == nil {
		t.Error("Expected publishedProductId output")
	}
}

// TestCSharpCollectionItemExample tests the C# CollectionItem example
func TestCSharpCollectionItemExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("collectionitem", "csharp"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedCollectionId"].Value == nil {
		t.Error("Expected deployedCollectionId output")
	}
	if result.Outputs["draftPostId"].Value == nil {
		t.Error("Expected draftPostId output")
	}
	if result.Outputs["publishedProductId"].Value == nil {
		t.Error("Expected publishedProductId output")
	}
}

// TestJavaCollectionItemExample tests the Java CollectionItem example
func TestJavaCollectionItemExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("collectionitem", "java"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedCollectionId"].Value == nil {
		t.Error("Expected deployedCollectionId output")
	}
	if result.Outputs["draftPostId"].Value == nil {
		t.Error("Expected draftPostId output")
	}
	if result.Outputs["publishedProductId"].Value == nil {
		t.Error("Expected publishedProductId output")
	}
}
