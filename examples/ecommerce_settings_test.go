//go:build all || typescript
// +build all typescript

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptEcommerceSettingsExample tests the TypeScript EcommerceSettings example
func TestTypeScriptEcommerceSettingsExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("ecommerce-settings", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	// Set required config - siteId is required by the EcommerceSettings resource
	// Note: This site must have ecommerce enabled for the test to pass
	test.SetConfig(t, "siteId", "580e63e98c9a982ac9b8b741")

	test.Preview(t)
	result := test.Up(t)

	// Validate ecommerce settings outputs
	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
	if result.Outputs["ecommerceSiteId"].Value == nil {
		t.Error("Expected ecommerceSiteId output")
	}
	if result.Outputs["defaultCurrency"].Value == nil {
		t.Error("Expected defaultCurrency output")
	}
	if result.Outputs["ecommerceCreatedOn"].Value == nil {
		t.Error("Expected ecommerceCreatedOn output")
	}

	// Validate currency format (should be 3-letter ISO code)
	currency, ok := result.Outputs["defaultCurrency"].Value.(string)
	if ok && len(currency) != 3 {
		t.Errorf("Expected defaultCurrency to be 3-letter ISO code, got: %s", currency)
	}
}
