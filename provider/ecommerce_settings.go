// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

// EcommerceSettingsResponse represents the Webflow API response for ecommerce settings.
// This struct matches the GET /v2/sites/{site_id}/ecommerce/settings endpoint response.
type EcommerceSettingsResponse struct {
	// SiteID is the identifier of the Site.
	SiteID string `json:"siteId"`
	// CreatedOn is the date when the ecommerce settings were created (ISO 8601 format).
	CreatedOn string `json:"createdOn"`
	// DefaultCurrency is the three-letter ISO currency code for the Site (e.g., "USD", "EUR").
	DefaultCurrency string `json:"defaultCurrency"`
}

// currencyCodePattern is the regex pattern for validating ISO 4217 currency codes.
// Currency codes are 3-letter uppercase alphabetic codes.
var currencyCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)

// ValidateCurrencyCode validates that a currency code is a valid 3-letter ISO 4217 code.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateCurrencyCode(code string) error {
	if code == "" {
		return errors.New("defaultCurrency is required but was not provided. " +
			"Please provide a valid 3-letter ISO 4217 currency code (e.g., 'USD', 'EUR', 'GBP'). " +
			"You can find a full list of currency codes at https://www.iso.org/iso-4217-currency-codes.html")
	}
	if !currencyCodePattern.MatchString(code) {
		return fmt.Errorf("defaultCurrency has invalid format: got '%s'. "+
			"Expected a 3-letter uppercase ISO 4217 currency code (e.g., 'USD', 'EUR', 'GBP'). "+
			"Currency codes must be exactly 3 uppercase letters. "+
			"Common codes: USD (US Dollar), EUR (Euro), GBP (British Pound), JPY (Japanese Yen)", code)
	}
	return nil
}

// GenerateEcommerceSettingsResourceID generates a Pulumi resource ID for an EcommerceSettings resource.
// Format: {siteID}/ecommerce/settings
// Note: EcommerceSettings is a 1:1 relationship with a site.
func GenerateEcommerceSettingsResourceID(siteID string) string {
	return siteID + "/ecommerce/settings"
}

// ExtractSiteIDFromEcommerceSettingsResourceID extracts the siteID from an EcommerceSettings resource ID.
// Expected format: {siteID}/ecommerce/settings
func ExtractSiteIDFromEcommerceSettingsResourceID(resourceID string) (string, error) {
	if resourceID == "" {
		return "", errors.New("resourceId cannot be empty")
	}

	suffix := "/ecommerce/settings"
	if len(resourceID) <= len(suffix) {
		return "", fmt.Errorf("invalid resource ID format: expected {siteId}/ecommerce/settings, got: %s", resourceID)
	}

	if resourceID[len(resourceID)-len(suffix):] != suffix {
		return "", fmt.Errorf("invalid resource ID format: expected {siteId}/ecommerce/settings, got: %s", resourceID)
	}

	siteID := resourceID[:len(resourceID)-len(suffix)]
	if siteID == "" {
		return "", fmt.Errorf("invalid resource ID format: expected {siteId}/ecommerce/settings, got: %s", resourceID)
	}

	return siteID, nil
}

// getEcommerceSettingsBaseURL is used internally for testing to override the API base URL.
var getEcommerceSettingsBaseURL = ""

// GetEcommerceSettings retrieves the ecommerce settings for a Webflow site.
// It calls GET /v2/sites/{site_id}/ecommerce/settings endpoint.
// Returns the parsed response or an error if the request fails.
//
// Note: This endpoint requires the ecommerce:read scope and will return a 409 error
// if ecommerce is not enabled on the site.
func GetEcommerceSettings(ctx context.Context, client *http.Client, siteID string) (*EcommerceSettingsResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getEcommerceSettingsBaseURL != "" {
		baseURL = getEcommerceSettingsBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/ecommerce/settings", baseURL, siteID)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = handleNetworkError(err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close() // Close immediately after reading
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Handle rate limiting with retry
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			var waitTime time.Duration
			if retryAfter != "" {
				waitTime = getRetryAfterDuration(retryAfter, time.Duration(1<<uint(attempt))*time.Second)
			} else {
				waitTime = time.Duration(1<<uint(attempt)) * time.Second
			}

			// Enhanced rate limiting error message with clear delay information
			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support",
				attempt+1, maxRetries+1, waitTime)

			// Wait before next retry if we haven't exhausted retries
			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(waitTime):
				}
			}
			continue
		}

		// Handle 409 Conflict - Ecommerce not enabled
		if resp.StatusCode == 409 {
			return nil, handleEcommerceNotEnabledError(body)
		}

		// Handle error responses
		if resp.StatusCode != 200 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var response EcommerceSettingsResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// handleEcommerceNotEnabledError creates an actionable error message when ecommerce is not enabled.
func handleEcommerceNotEnabledError(body []byte) error {
	return fmt.Errorf("ecommerce not enabled: the site does not have ecommerce enabled. "+
		"To fix this: 1) Log into your Webflow dashboard, 2) Go to Site Settings > Ecommerce, "+
		"3) Enable ecommerce for this site, 4) Set up your payment provider and currency settings, "+
		"5) Retry this operation. Details: %s", string(body))
}
