//go:build nodejs || all
// +build nodejs all

package examples

import (
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

func TestNodejsExampleLifecycle(t *testing.T) {
	pt := pulumitest.NewPulumiTest(t, "nodejs",
		opttest.YarnLink("@jdetmar/pulumi-webflow"),
		opttest.AttachProviderServer("webflow", providerFactory),
		opttest.Env("PULUMI_PREFER_YARN", "true"),
	)

	pt.Preview(t)
	pt.Up(t)
	pt.Destroy(t)
}
