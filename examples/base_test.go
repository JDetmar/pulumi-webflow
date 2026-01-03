package examples

import (
	"os"
	"testing"

	"github.com/jdetmar/pulumi-webflow/provider"
	"github.com/pulumi/providertest/providers"
	goprovider "github.com/pulumi/pulumi-go-provider"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

var providerFactory = func(_ providers.PulumiTest) (pulumirpc.ResourceProviderServer, error) {
	return goprovider.RawServer("webflow", "0.1.0", provider.Provider())(nil)
}

// skipIfNoAPIToken skips the test if WEBFLOW_API_TOKEN is not set.
// This allows tests to pass in CI environments without valid credentials.
func skipIfNoAPIToken(t *testing.T) {
	t.Helper()
	if os.Getenv("WEBFLOW_API_TOKEN") == "" {
		t.Skip("Skipping: WEBFLOW_API_TOKEN not set. Set this environment variable to run integration tests.")
	}
}
