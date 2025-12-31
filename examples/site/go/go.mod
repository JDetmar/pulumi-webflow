module github.com/jdetmar/pulumi-webflow/examples/site/go

go 1.21

require (
	github.com/jdetmar/pulumi-webflow/sdk/go/webflow v0.0.0
	github.com/pulumi/pulumi/sdk/v3 v3.0.0
)

replace github.com/jdetmar/pulumi-webflow/sdk/go/webflow => ../../../sdk/go/webflow
