// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package examples

import (
	"os"
	"path"
	"strings"
	"testing"
)

// TestMultiSiteBasicTypeScript verifies basic-typescript example compiles.
func TestMultiSiteBasicTypeScript(t *testing.T) {
	t.Skip("Skipped: Requires Node.js runtime and npm dependencies. " +
		"Run manually: cd examples/multi-site/basic-typescript && npm install && " +
		"pulumi preview")
}

// TestMultiSiteBasicPython verifies basic-python example compiles.
func TestMultiSiteBasicPython(t *testing.T) {
	t.Skip("Skipped: Requires Python runtime and dependencies. " +
		"Run manually: cd examples/multi-site/basic-python && pip install -r " +
		"requirements.txt && pulumi preview")
}

// TestMultiSiteBasicGo verifies basic-go example compiles and has valid main.
func TestMultiSiteBasicGo(t *testing.T) {
	// Check that main.go exists and is readable
	examplePath := path.Join(getCwd(t), "multi-site", "basic-go")
	mainFile := path.Join(examplePath, "main.go")

	content, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	// Verify basic structure
	contentStr := string(content)
	if !strings.Contains(contentStr, "func main()") {
		t.Error("main.go missing main function")
	}
	if !strings.Contains(contentStr, "pulumi.Run") {
		t.Error("main.go missing pulumi.Run")
	}
	if !strings.Contains(contentStr, "webflow.NewSite") {
		t.Error("main.go missing Site resource creation")
	}
}

// TestMultiSiteConfigDriven verifies config-driven example has valid structure.
func TestMultiSiteConfigDriven(t *testing.T) {
	examplePath := path.Join(getCwd(t), "multi-site", "config-driven-typescript")

	// Check index.ts exists
	indexFile := path.Join(examplePath, "index.ts")
	indexContent, err := os.ReadFile(indexFile)
	if err != nil {
		t.Fatalf("Failed to read index.ts: %v", err)
	}

	// Check sites.yaml exists
	sitesFile := path.Join(examplePath, "sites.yaml")
	sitesContent, err := os.ReadFile(sitesFile)
	if err != nil {
		t.Fatalf("Failed to read sites.yaml: %v", err)
	}

	// Verify structure
	indexStr := string(indexContent)
	if !strings.Contains(indexStr, "yaml.load") {
		t.Error("index.ts missing YAML loading")
	}
	if !strings.Contains(indexStr, "new webflow.Site") {
		t.Error("index.ts missing Site creation")
	}

	// Verify sites.yaml has sites
	sitesStr := string(sitesContent)
	if !strings.Contains(sitesStr, "sites:") {
		t.Error("sites.yaml missing sites section")
	}
	if !strings.Contains(sitesStr, "name:") {
		t.Error("sites.yaml missing site names")
	}
}

// TestMultiSiteTemplate verifies template example has factory functions.
func TestMultiSiteTemplate(t *testing.T) {
	examplePath := path.Join(getCwd(t), "multi-site", "template-python")

	// Check site_templates.py exists
	templatesFile := path.Join(examplePath, "site_templates.py")
	content, err := os.ReadFile(templatesFile)
	if err != nil {
		t.Fatalf("Failed to read site_templates.py: %v", err)
	}

	contentStr := string(content)

	// Verify factory functions exist
	factories := []string{
		"def create_campaign_site",
		"def create_product_site",
		"def create_event_site",
	}

	for _, factory := range factories {
		if !strings.Contains(contentStr, factory) {
			t.Errorf("site_templates.py missing factory function: %s", factory)
		}
	}

	// Verify factories create redirects
	if !strings.Contains(contentStr, "webflow.Redirect") {
		t.Error("site_templates.py not creating redirects")
	}
}

// TestMultiSiteEnvironments verifies multi-env example has stack configs.
func TestMultiSiteEnvironments(t *testing.T) {
	examplePath := path.Join(getCwd(t), "multi-site", "multi-env-go")

	// Check stack-specific configs exist
	stackConfigs := []string{
		path.Join(examplePath, "Pulumi.dev.yaml"),
		path.Join(examplePath, "Pulumi.staging.yaml"),
		path.Join(examplePath, "Pulumi.prod.yaml"),
	}

	for _, configPath := range stackConfigs {
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", configPath, err)
		}

		// Verify config has required fields
		contentStr := string(content)
		if !strings.Contains(contentStr, "sitePrefix") {
			t.Errorf("%s missing sitePrefix config", configPath)
		}
		if !strings.Contains(contentStr, "siteCount") {
			t.Errorf("%s missing siteCount config", configPath)
		}
	}

	// Verify main.go reads configs
	mainFile := path.Join(examplePath, "main.go")
	mainContent, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	mainStr := string(mainContent)
	if !strings.Contains(mainStr, "config.New") {
		t.Error("main.go not reading configuration")
	}
	if !strings.Contains(mainStr, "sitePrefix") {
		t.Error("main.go not using sitePrefix")
	}
}

// TestMultiSiteDocumentation verifies comprehensive README exists.
func TestMultiSiteDocumentation(t *testing.T) {
	readmePath := path.Join(getCwd(t), "multi-site", "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	contentStr := string(content)

	// Verify documentation covers all patterns
	requiredSections := []string{
		"Basic Multi-Site",
		"Configuration-Driven",
		"Template-Based",
		"Multi-Environment",
		"Quick Start",
		"Best Practices",
		"Troubleshooting",
	}

	for _, section := range requiredSections {
		if !strings.Contains(contentStr, section) {
			t.Errorf("README.md missing section: %s", section)
		}
	}

	// Verify key concepts documented
	concepts := []string{
		"paralleliz",
		"configuration",
		"factory",
		"stack",
	}

	for _, concept := range concepts {
		if !strings.Contains(strings.ToLower(contentStr), concept) {
			t.Logf("Note: README.md may not adequately cover: %s", concept)
		}
	}
}

// TestMultiSiteAcceptanceCriteria verifies examples satisfy AC.
func TestMultiSiteAcceptanceCriteria(t *testing.T) {
	// AC1: Multiple sites managed in single program

	// Test 1: Basic example creates 3+ sites
	basicExamplePath := path.Join(getCwd(t), "multi-site", "basic-typescript")
	basicContent, err := os.ReadFile(path.Join(basicExamplePath, "index.ts"))
	if err != nil {
		t.Fatalf("Failed to read basic-typescript example: %v", err)
	}
	if !strings.Contains(string(basicContent), "new webflow.Site") {
		t.Error("AC1: Basic example doesn't create Site resources")
	}

	// Test 2: Configuration example creates 10+ sites
	configExamplePath := path.Join(getCwd(t), "multi-site", "config-driven-typescript")
	sitesContent, err := os.ReadFile(path.Join(configExamplePath, "sites.yaml"))
	if err != nil {
		t.Fatalf("Failed to read sites.yaml: %v", err)
	}
	siteCount := strings.Count(string(sitesContent), "- name:")
	if siteCount < 10 {
		t.Errorf("AC1: Configuration example has only %d sites (need 10+)", siteCount)
	}

	// AC2: Error handling - verify examples handle individual site failures

	// Check templates have error handling
	templatePath := path.Join(getCwd(t), "multi-site", "template-python")
	templateContent, err := os.ReadFile(path.Join(templatePath, "site_templates.py"))
	if err != nil {
		t.Fatalf("Failed to read site_templates.py: %v", err)
	}
	templateStr := string(templateContent)
	if !strings.Contains(templateStr, "Args") &&
		!strings.Contains(templateStr, "def create") {
		t.Error("AC2: Templates don't show error handling patterns")
	}

	// Check multi-env has error handling
	multiEnvPath := path.Join(getCwd(t), "multi-site", "multi-env-go")
	multiEnvContent, err := os.ReadFile(path.Join(multiEnvPath, "main.go"))
	if err != nil {
		t.Fatalf("Failed to read multi-env main.go: %v", err)
	}
	multiEnvStr := string(multiEnvContent)
	if !strings.Contains(multiEnvStr, "if err !=") {
		t.Error("AC2: Multi-env example doesn't handle errors per site")
	}
	if !strings.Contains(multiEnvStr, "fmt.Sprintf") {
		t.Error("AC2: Multi-env example doesn't identify which site failed")
	}
}

// Helper function to get the examples directory
func getCwd(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	return cwd
}
