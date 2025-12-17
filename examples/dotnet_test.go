//go:build dotnet || all
// +build dotnet all

package examples

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

func TestDotnet(t *testing.T) {
	pt := pulumitest.NewPulumiTest(t, "dotnet",
		opttest.DotNetReference("Pulumi.ProviderWebflow", "../sdk/dotnet"),
		opttest.AttachProviderServer("webflow", providerFactory),
	)

	pt.Preview(t)
	pt.Up(t)
	pt.Destroy(t)
}
