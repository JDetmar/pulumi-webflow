// Copyright 2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package examples

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStackConfigTypeScriptStructure(t *testing.T) {
	// Test that TypeScript example has all required files
	exampleDir := filepath.Join("stack-config", "typescript-complete")

	requiredFiles := []string{
		"index.ts",
		"Pulumi.yaml",
		"Pulumi.dev.yaml",
		"Pulumi.staging.yaml",
		"Pulumi.prod.yaml",
		"package.json",
		"tsconfig.json",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(exampleDir, file)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("TypeScript example missing file: %s", file)
		}
	}
}

func TestStackConfigPythonStructure(t *testing.T) {
	// Test that Python example has all required files
	exampleDir := filepath.Join("stack-config", "python-workflow")

	requiredFiles := []string{
		"__main__.py",
		"Pulumi.yaml",
		"Pulumi.dev.yaml",
		"Pulumi.staging.yaml",
		"Pulumi.prod.yaml",
		"requirements.txt",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(exampleDir, file)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("Python example missing file: %s", file)
		}
	}
}

func TestStackConfigGoStructure(t *testing.T) {
	// Test that Go example has all required files
	exampleDir := filepath.Join("stack-config", "go-advanced")

	requiredFiles := []string{
		"main.go",
		"Pulumi.yaml",
		"Pulumi.dev.yaml",
		"Pulumi.staging.yaml",
		"Pulumi.prod.yaml",
		"go.mod",
	}

	for _, file := range requiredFiles {
		filePath := filepath.Join(exampleDir, file)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("Go example missing file: %s", file)
		}
	}
}

func TestStackConfigREADME(t *testing.T) {
	// Test that comprehensive documentation exists
	readmePath := filepath.Join("stack-config", "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		t.Error("Stack configuration README.md not found")
	}
}

func TestStackConfigDevStackFiles(t *testing.T) {
	// Test that all examples have dev stack configuration
	examples := []string{
		filepath.Join("stack-config", "typescript-complete"),
		filepath.Join("stack-config", "python-workflow"),
		filepath.Join("stack-config", "go-advanced"),
	}

	for _, exampleDir := range examples {
		devFile := filepath.Join(exampleDir, "Pulumi.dev.yaml")
		if _, err := os.Stat(devFile); err != nil {
			t.Errorf("Dev stack configuration missing in %s", exampleDir)
		}
	}
}

func TestStackConfigStagingStackFiles(t *testing.T) {
	// Test that all examples have staging stack configuration
	examples := []string{
		filepath.Join("stack-config", "typescript-complete"),
		filepath.Join("stack-config", "python-workflow"),
		filepath.Join("stack-config", "go-advanced"),
	}

	for _, exampleDir := range examples {
		stagingFile := filepath.Join(exampleDir, "Pulumi.staging.yaml")
		if _, err := os.Stat(stagingFile); err != nil {
			t.Errorf("Staging stack configuration missing in %s", exampleDir)
		}
	}
}

func TestStackConfigProdStackFiles(t *testing.T) {
	// Test that all examples have production stack configuration
	examples := []string{
		filepath.Join("stack-config", "typescript-complete"),
		filepath.Join("stack-config", "python-workflow"),
		filepath.Join("stack-config", "go-advanced"),
	}

	for _, exampleDir := range examples {
		prodFile := filepath.Join(exampleDir, "Pulumi.prod.yaml")
		if _, err := os.Stat(prodFile); err != nil {
			t.Errorf("Production stack configuration missing in %s", exampleDir)
		}
	}
}

func TestStackConfigProjFiles(t *testing.T) {
	// Test that all examples have project definition
	examples := []string{
		filepath.Join("stack-config", "typescript-complete"),
		filepath.Join("stack-config", "python-workflow"),
		filepath.Join("stack-config", "go-advanced"),
	}

	for _, exampleDir := range examples {
		projFile := filepath.Join(exampleDir, "Pulumi.yaml")
		if _, err := os.Stat(projFile); err != nil {
			t.Errorf("Project definition missing in %s", exampleDir)
		}
	}
}

func TestStackConfigDependencies(t *testing.T) {
	// Test that TypeScript example has package.json with dependencies
	packageFile := filepath.Join("stack-config", "typescript-complete", "package.json")
	if _, err := os.Stat(packageFile); err != nil {
		t.Error("TypeScript package.json not found")
	}

	// Test that Python example has requirements.txt
	requirementsFile := filepath.Join("stack-config", "python-workflow", "requirements.txt")
	if _, err := os.Stat(requirementsFile); err != nil {
		t.Error("Python requirements.txt not found")
	}

	// Test that Go example has go.mod
	gomodFile := filepath.Join("stack-config", "go-advanced", "go.mod")
	if _, err := os.Stat(gomodFile); err != nil {
		t.Error("Go go.mod not found")
	}
}

func TestStackConfigEntryPoints(t *testing.T) {
	// Test that each example has correct entry point file
	entryPoints := map[string]string{
		"stack-config/typescript-complete": "index.ts",
		"stack-config/python-workflow":     "__main__.py",
		"stack-config/go-advanced":         "main.go",
	}

	for exampleDir, entryPoint := range entryPoints {
		filePath := filepath.Join(exampleDir, entryPoint)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("%s missing entry point: %s", exampleDir, entryPoint)
		}
	}
}

func TestStackConfigSecurityPatterns(t *testing.T) {
	// Verify that production stack configs have safety checks
	prodConfigPath := filepath.Join("stack-config", "typescript-complete", "Pulumi.prod.yaml")
	prodContent, err := os.ReadFile(prodConfigPath) //nolint:gosec // G304: Test file with controlled path
	if err != nil {
		t.Fatalf("Could not read prod config: %v", err)
	}

	// Check for production safety requirement
	if !strings.Contains(string(prodContent), "prodDeploymentConfirmed") {
		t.Error("Production config missing prodDeploymentConfirmed safety check")
	}

	if !strings.Contains(string(prodContent), "environmentName: prod") {
		t.Error("Production config missing environmentName")
	}
}

func TestStackConfigEnvironmentSpecification(t *testing.T) {
	// Verify that each stack config specifies its environment
	configs := map[string]string{
		"stack-config/typescript-complete/Pulumi.dev.yaml":     "dev",
		"stack-config/typescript-complete/Pulumi.staging.yaml": "staging",
		"stack-config/typescript-complete/Pulumi.prod.yaml":    "prod",
		"stack-config/python-workflow/Pulumi.dev.yaml":         "dev",
		"stack-config/python-workflow/Pulumi.staging.yaml":     "staging",
		"stack-config/python-workflow/Pulumi.prod.yaml":        "prod",
		"stack-config/go-advanced/Pulumi.dev.yaml":             "dev",
		"stack-config/go-advanced/Pulumi.staging.yaml":         "staging",
		"stack-config/go-advanced/Pulumi.prod.yaml":            "prod",
	}

	for configPath, expectedEnv := range configs {
		content, err := os.ReadFile(configPath) //nolint:gosec // G304: Test file with controlled path
		if err != nil {
			t.Fatalf("Could not read config %s: %v", configPath, err)
		}

		expectedStr := "environmentName: " + expectedEnv
		if !strings.Contains(string(content), expectedStr) {
			t.Errorf("%s missing environmentName: %s", configPath, expectedEnv)
		}
	}
}

func TestStackConfigMultipleStacksPerExample(t *testing.T) {
	// Verify that each example supports dev, staging, and prod stacks
	examples := []string{
		"stack-config/typescript-complete",
		"stack-config/python-workflow",
		"stack-config/go-advanced",
	}

	stackNames := []string{"dev", "staging", "prod"}

	for _, example := range examples {
		for _, stack := range stackNames {
			stackFile := filepath.Join(example, "Pulumi."+stack+".yaml")
			if _, err := os.Stat(stackFile); err != nil {
				t.Errorf("Example %s missing stack configuration: %s", example, stack)
			}
		}
	}
}

func TestStackConfigAcceptanceCriteria(t *testing.T) {
	// Verify implementation meets Story 5.3 acceptance criteria

	// AC1: Multiple stacks use stack-specific configuration
	// Verify all examples have dev/staging/prod stacks
	examples := []string{
		"stack-config/typescript-complete",
		"stack-config/python-workflow",
		"stack-config/go-advanced",
	}

	for _, example := range examples {
		for _, stack := range []string{"dev", "staging", "prod"} {
			stackFile := filepath.Join(example, "Pulumi."+stack+".yaml")
			if _, err := os.Stat(stackFile); err != nil {
				t.Errorf("AC1 FAIL: %s missing %s stack", example, stack)
			}
		}
	}

	// AC2: Verify production safety checks exist
	prodFiles := []string{
		"stack-config/typescript-complete/Pulumi.prod.yaml",
		"stack-config/python-workflow/Pulumi.prod.yaml",
		"stack-config/go-advanced/Pulumi.prod.yaml",
	}

	for _, prodFile := range prodFiles {
		content, err := os.ReadFile(prodFile) //nolint:gosec // G304: Test file with controlled path
		if err != nil {
			t.Fatalf("Could not read %s: %v", prodFile, err)
		}

		if !strings.Contains(string(content), "prod") {
			t.Errorf("AC2 FAIL: %s missing prod environment specification", prodFile)
		}
	}

	// Verify README provides stack configuration guidance
	readmePath := filepath.Join("stack-config", "README.md")
	readmeContent, err := os.ReadFile(readmePath) //nolint:gosec // G304: Test file with controlled path
	if err != nil {
		t.Fatalf("Could not read README: %v", err)
	}

	if !strings.Contains(string(readmeContent), "Stack Configuration") {
		t.Error("AC1 FAIL: README missing Stack Configuration section")
	}

	if !strings.Contains(string(readmeContent), "Credential") {
		t.Error("AC2 FAIL: README missing Credential Management section")
	}
}

func TestStackConfigSitesPattern(t *testing.T) {
	// Verify that stack configs use the realistic sites pattern (not siteCount)
	configs := []string{
		"stack-config/typescript-complete/Pulumi.dev.yaml",
		"stack-config/python-workflow/Pulumi.dev.yaml",
		"stack-config/go-advanced/Pulumi.dev.yaml",
	}

	for _, configPath := range configs {
		content, err := os.ReadFile(configPath) //nolint:gosec // G304: Test file with controlled path
		if err != nil {
			t.Fatalf("Could not read %s: %v", configPath, err)
		}

		// Should have sites configuration (realistic pattern)
		if !strings.Contains(string(content), "sites:") {
			t.Errorf("%s missing sites configuration", configPath)
		}

		// Should have named sites like marketing, docs, blog
		if !strings.Contains(string(content), "marketing:") {
			t.Errorf("%s missing marketing site definition", configPath)
		}
	}
}
