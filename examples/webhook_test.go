//go:build all || typescript || python || go
// +build all typescript python go

package examples

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestTypeScriptWebhookExample tests the TypeScript Webhook example
func TestTypeScriptWebhookExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("webhook", "typescript"),
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
	if result.Outputs["formWebhookId"].Value == nil {
		t.Error("Expected formWebhookId output")
	}
}

// TestPythonWebhookExample tests the Python Webhook example
func TestPythonWebhookExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("webhook", "python"),
		opttest.PythonLink("../sdk/python"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployed_site_id"].Value == nil {
		t.Error("Expected deployed_site_id output")
	}
	if result.Outputs["form_webhook_id"].Value == nil {
		t.Error("Expected form_webhook_id output")
	}
}

// TestGoWebhookExample tests the Go Webhook example
func TestGoWebhookExample(t *testing.T) {
	skipIfNoAPIToken(t)

	test := pulumitest.NewPulumiTest(t,
		filepath.Join("webhook", "go"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	test.Preview(t)
	result := test.Up(t)

	if result.Outputs["deployedSiteId"].Value == nil {
		t.Error("Expected deployedSiteId output")
	}
	if result.Outputs["formWebhookId"].Value == nil {
		t.Error("Expected formWebhookId output")
	}
}
