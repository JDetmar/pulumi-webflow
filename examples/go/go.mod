module github.com/jdetmar/pulumi-webflow/examples/go

go 1.24

require (
	github.com/jdetmar/pulumi-webflow/sdk/go v0.0.0
	github.com/pulumi/pulumi/sdk/v3 v3.212.0
)

replace github.com/jdetmar/pulumi-webflow/sdk/go => ../../sdk/go
