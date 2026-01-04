// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package examples

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoggingExamplesStructure validates the directory and file structure
func TestLoggingExamplesStructure(t *testing.T) {
	baseDir := "troubleshooting-logs"

	// Verify main directory exists
	_, err := os.Stat(baseDir)
	assert.NoError(t, err, "troubleshooting-logs directory should exist")

	// Expected subdirectories
	subdirs := []string{
		"typescript-troubleshooting",
		"python-cicd-logging",
		"go-log-analysis",
	}

	for _, subdir := range subdirs {
		path := filepath.Join(baseDir, subdir)
		_, err := os.Stat(path)
		assert.NoError(t, err, "subdirectory %s should exist", subdir)
	}

	// Verify README exists
	readmePath := filepath.Join(baseDir, "README.md")
	_, err = os.Stat(readmePath)
	assert.NoError(t, err, "README.md should exist")
}

// TestTypeScriptLoggingExample validates TypeScript example structure
func TestTypeScriptLoggingExample(t *testing.T) {
	exampleDir := "troubleshooting-logs/typescript-troubleshooting"

	requiredFiles := []string{
		"index.ts",
		"Pulumi.yaml",
		"package.json",
		"tsconfig.json",
		".gitignore",
	}

	for _, file := range requiredFiles {
		path := filepath.Join(exampleDir, file)
		_, err := os.Stat(path)
		assert.NoError(t, err, "%s should exist", file)
	}

	// Verify index.ts contains logging calls
	indexPath := filepath.Join(exampleDir, "index.ts")
	content, err := os.ReadFile(indexPath)
	require.NoError(t, err)

	indexStr := string(content)
	assert.Contains(t, indexStr, "pulumi.log.info", "index.ts should contain info logging")
	assert.Contains(t, indexStr, "pulumi.log.debug", "index.ts should contain debug logging")
	assert.Contains(t, indexStr, "pulumi-webflow", "index.ts should import webflow SDK")

	// Verify copyright header
	assert.Contains(t, indexStr, "Copyright 2025, Pulumi Corporation", "should have Apache 2.0 copyright")
}

// TestPythonCICDLoggingExample validates Python example structure
func TestPythonCICDLoggingExample(t *testing.T) {
	exampleDir := "troubleshooting-logs/python-cicd-logging"

	requiredFiles := []string{
		"__main__.py",
		"Pulumi.yaml",
		"requirements.txt",
		".gitignore",
	}

	for _, file := range requiredFiles {
		path := filepath.Join(exampleDir, file)
		_, err := os.Stat(path)
		assert.NoError(t, err, "%s should exist", file)
	}

	// Verify __main__.py contains CI/CD patterns
	mainPath := filepath.Join(exampleDir, "__main__.py")
	content, err := os.ReadFile(mainPath)
	require.NoError(t, err)

	mainStr := string(content)
	assert.Contains(t, mainStr, "pulumi.log.info", "__main__.py should contain info logging")
	assert.Contains(t, mainStr, "os.getenv", "should detect environment variables")
	assert.Contains(t, mainStr, "CI", "should detect CI/CD environment")
	assert.Contains(t, mainStr, "PULUMI_STACK", "should detect Pulumi stack")

	// Verify copyright header
	assert.Contains(t, mainStr, "Copyright 2025, Pulumi Corporation", "should have Apache 2.0 copyright")
}

// TestGoLogAnalysisExample validates Go example structure
func TestGoLogAnalysisExample(t *testing.T) {
	exampleDir := "troubleshooting-logs/go-log-analysis"

	requiredFiles := []string{
		"main.go",
		"Pulumi.yaml",
		"go.mod",
		".gitignore",
	}

	for _, file := range requiredFiles {
		path := filepath.Join(exampleDir, file)
		_, err := os.Stat(path)
		assert.NoError(t, err, "%s should exist", file)
	}

	// Verify main.go contains logging patterns
	mainPath := filepath.Join(exampleDir, "main.go")
	content, err := os.ReadFile(mainPath)
	require.NoError(t, err)

	mainStr := string(content)
	assert.Contains(t, mainStr, "ctx.Log.Info", "main.go should contain info logging")
	assert.Contains(t, mainStr, "ctx.Log.Debug", "main.go should contain debug logging")
	assert.Contains(t, mainStr, "ctx.Log.Error", "main.go should contain error logging")

	// Verify copyright header
	assert.Contains(t, mainStr, "Copyright 2025, Pulumi Corporation", "should have Apache 2.0 copyright")

	// Verify Apache 2.0 license text
	assert.Contains(t, mainStr, "Apache License, Version 2.0", "should reference Apache 2.0 license")
}

// TestCredentialRedactionVerification validates credential redaction patterns
func TestCredentialRedactionVerification(t *testing.T) {
	// Verify auth.go contains RedactToken function
	authPath := "../provider/auth.go"
	content, err := os.ReadFile(authPath)
	require.NoError(t, err, "../provider/auth.go should exist for redaction verification")

	authStr := string(content)
	assert.Contains(t, authStr, "RedactToken", "auth.go should have RedactToken function")
	assert.Contains(t, authStr, "[REDACTED]", "redaction should use [REDACTED] placeholder")

	// Verify no example files contain actual token patterns
	exampleDirs := []string{
		"troubleshooting-logs/typescript-troubleshooting",
		"troubleshooting-logs/python-cicd-logging",
		"troubleshooting-logs/go-log-analysis",
	}

	for _, exampleDir := range exampleDirs {
		err := filepath.Walk(exampleDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".py") &&
				!strings.HasSuffix(path, ".go") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			fileStr := string(content)
			// No example should contain fake tokens starting with wf_
			assert.NotContains(t, fileStr, "wf_", "example files should not contain token patterns")
			return nil
		})

		assert.NoError(t, err, "should walk %s directory without errors", exampleDir)
	}
}

// TestLoggingConfigurationPatterns validates logging configuration approaches
func TestLoggingConfigurationPatterns(t *testing.T) {
	// Check TypeScript uses Pulumi logging API
	tsPath := "troubleshooting-logs/typescript-troubleshooting/index.ts"
	tsContent, err := os.ReadFile(tsPath)
	require.NoError(t, err)
	tsStr := string(tsContent)
	assert.Contains(t, tsStr, "pulumi.log", "TypeScript should use pulumi.log API")

	// Check Python uses Pulumi logging API
	pyPath := "troubleshooting-logs/python-cicd-logging/__main__.py"
	pyContent, err := os.ReadFile(pyPath)
	require.NoError(t, err)
	pyStr := string(pyContent)
	assert.Contains(t, pyStr, "pulumi.log", "Python should use pulumi.log API")

	// Check Go uses context logging
	goPath := "troubleshooting-logs/go-log-analysis/main.go"
	goContent, err := os.ReadFile(goPath)
	require.NoError(t, err)
	goStr := string(goContent)
	assert.Contains(t, goStr, "ctx.Log", "Go should use ctx.Log API")
}

// TestLoggingPerformanceGuidance validates performance guidance
func TestLoggingPerformanceGuidance(t *testing.T) {
	// Check README contains performance guidance
	readmePath := "troubleshooting-logs/README.md"
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err)

	readmeStr := string(content)
	assert.Contains(t, readmeStr, "Performance", "README should have performance section")
	assert.Contains(t, readmeStr, "production", "README should mention production considerations")
	assert.Contains(t, readmeStr, "verbose logging", "README should explain verbose logging impact")
	assert.Contains(t, readmeStr, "overhead", "README should quantify logging overhead")
}

// TestREADMEStructure validates the comprehensive README structure
func TestREADMEStructure(t *testing.T) {
	readmePath := "troubleshooting-logs/README.md"
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err)

	readmeStr := string(content)

	// Verify all 9 sections exist
	sections := []string{
		"Introduction",
		"Quick Start",
		"Pulumi Logging Levels",
		"Credential Redaction",
		"Common Troubleshooting Scenarios",
		"CI/CD Logging Configuration",
		"Log Analysis Techniques",
		"Performance Considerations",
		"Troubleshooting",
	}

	for _, section := range sections {
		assert.Contains(t, readmeStr, "## "+section, "README should have %s section", section)
	}

	// Verify content quality indicators
	assert.Contains(t, readmeStr, "```bash", "README should contain bash examples")
	assert.Contains(t, readmeStr, "```python", "README should contain Python examples")
	assert.Contains(t, readmeStr, "```go", "README should contain Go examples")
	assert.Contains(t, readmeStr, "| ", "README should contain tables")
	assert.Contains(t, readmeStr, "✅", "README should use helpful formatting")
	assert.Contains(t, readmeStr, "❌", "README should use helpful formatting")
}

// TestGitignoreFiles validates all examples have proper .gitignore
func TestGitignoreFiles(t *testing.T) {
	examples := []string{
		"troubleshooting-logs/typescript-troubleshooting/.gitignore",
		"troubleshooting-logs/python-cicd-logging/.gitignore",
		"troubleshooting-logs/go-log-analysis/.gitignore",
	}

	for _, gitignorePath := range examples {
		content, err := os.ReadFile(gitignorePath)
		assert.NoError(t, err, "%s should exist", gitignorePath)

		gitignoreStr := string(content)

		// Verify Pulumi backup files are ignored
		assert.Contains(t, gitignoreStr, "Pulumi.*.yaml.backup", "%s should ignore Pulumi backup files", gitignorePath)

		// Verify build artifacts are ignored
		assert.Contains(t, gitignoreStr, ".DS_Store", "%s should ignore OS files", gitignorePath)
	}
}

// TestPulumiConfigFiles validates Pulumi.yaml files
func TestPulumiConfigFiles(t *testing.T) {
	configs := []string{
		"troubleshooting-logs/typescript-troubleshooting/Pulumi.yaml",
		"troubleshooting-logs/python-cicd-logging/Pulumi.yaml",
		"troubleshooting-logs/go-log-analysis/Pulumi.yaml",
	}

	for _, configPath := range configs {
		content, err := os.ReadFile(configPath)
		assert.NoError(t, err, "%s should exist", configPath)

		configStr := string(content)
		assert.Contains(t, configStr, "name:", "%s should have project name", configPath)
		assert.Contains(t, configStr, "runtime:", "%s should specify runtime", configPath)
		assert.Contains(t, configStr, "description:", "%s should have description", configPath)
	}
}
