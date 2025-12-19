package examples

import (
	"github.com/pulumi/providertest/providers"
	goprovider "github.com/pulumi/pulumi-go-provider"
	"github.com/jdetmar/pulumi-webflow/provider"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

var providerFactory = func(_ providers.PulumiTest) (pulumirpc.ResourceProviderServer, error) {
	return goprovider.RawServer("webflow", "0.1.0", provider.Provider())(nil)
}
