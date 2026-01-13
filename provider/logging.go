// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"fmt"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
)

// LogContext provides structured logging with consistent field formatting.
// It wraps Pulumi's logger with helpers for common logging patterns.
type LogContext struct {
	logger p.Logger
	fields map[string]interface{}
}

// NewLogContext creates a new logging context from the Pulumi provider context.
func NewLogContext(ctx context.Context) *LogContext {
	return &LogContext{
		logger: p.GetLogger(ctx),
		fields: make(map[string]interface{}),
	}
}

// WithField adds a field to the log context (chainable).
func (lc *LogContext) WithField(key string, value interface{}) *LogContext {
	lc.fields[key] = value
	return lc
}

// WithFields adds multiple fields to the log context (chainable).
func (lc *LogContext) WithFields(fields map[string]interface{}) *LogContext {
	for k, v := range fields {
		lc.fields[k] = v
	}
	return lc
}

// formatMessage formats a message with structured fields.
func (lc *LogContext) formatMessage(msg string) string {
	if len(lc.fields) == 0 {
		return msg
	}

	var parts []string
	for k, v := range lc.fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return fmt.Sprintf("%s [%s]", msg, strings.Join(parts, ", "))
}

// Debug logs a debug-level message with structured fields.
// Use for detailed information useful during development and debugging.
func (lc *LogContext) Debug(msg string) {
	lc.logger.Debug(lc.formatMessage(msg))
}

// Debugf logs a debug-level formatted message with structured fields.
func (lc *LogContext) Debugf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	lc.logger.Debug(lc.formatMessage(msg))
}

// Info logs an info-level message with structured fields.
// Use for general operational information.
func (lc *LogContext) Info(msg string) {
	lc.logger.Info(lc.formatMessage(msg))
}

// Infof logs an info-level formatted message with structured fields.
func (lc *LogContext) Infof(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	lc.logger.Info(lc.formatMessage(msg))
}

// Warn logs a warning-level message with structured fields.
// Use for potentially problematic situations that don't prevent operation.
func (lc *LogContext) Warn(msg string) {
	lc.logger.Warning(lc.formatMessage(msg))
}

// Warnf logs a warning-level formatted message with structured fields.
func (lc *LogContext) Warnf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	lc.logger.Warning(lc.formatMessage(msg))
}

// Error logs an error-level message with structured fields.
// Use for errors that prevent an operation from completing.
func (lc *LogContext) Error(msg string) {
	lc.logger.Error(lc.formatMessage(msg))
}

// Errorf logs an error-level formatted message with structured fields.
func (lc *LogContext) Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	lc.logger.Error(lc.formatMessage(msg))
}

// RedactSensitiveData redacts sensitive information from a string value.
// This should be used for any data that might contain tokens, passwords, or PII.
func RedactSensitiveData(value string) string {
	if value == "" {
		return "<empty>"
	}
	// For tokens and sensitive data, completely redact
	if len(value) > 40 {
		// Long values (likely tokens) - show first/last 4 chars
		return value[:4] + "..." + value[len(value)-4:]
	}
	// Short values - just redact
	return "[REDACTED]"
}

// TruncateForLogging truncates large strings to prevent log spam.
// Useful for response bodies and large payloads.
func TruncateForLogging(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + fmt.Sprintf("... (truncated, %d total chars)", len(value))
}

// SafeString converts any value to a string safely for logging.
// Handles nil values and redacts sensitive data based on field name.
func SafeString(fieldName string, value interface{}) string {
	if value == nil {
		return "<nil>"
	}

	str := fmt.Sprintf("%v", value)

	// Check if field name suggests sensitive data
	lowerName := strings.ToLower(fieldName)
	if strings.Contains(lowerName, "token") ||
		strings.Contains(lowerName, "password") ||
		strings.Contains(lowerName, "secret") ||
		strings.Contains(lowerName, "key") ||
		strings.Contains(lowerName, "authorization") {
		return RedactToken(str)
	}

	return str
}
