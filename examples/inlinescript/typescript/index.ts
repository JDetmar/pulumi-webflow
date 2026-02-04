import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const siteId = config.requireSecret("siteId");

/**
 * InlineScript Example - Registering Inline Custom Code Scripts
 *
 * This example demonstrates how to register inline JavaScript code snippets
 * directly in your Webflow site's script registry. Unlike RegisteredScript
 * (which references externally hosted files), InlineScript embeds the source
 * code directly. Registered inline scripts can then be deployed across your
 * site using the SiteCustomCode or PageCustomCode resources.
 *
 * Inline scripts must be:
 * - 2000 characters or fewer
 * - Follow semantic versioning
 */

// Example 1: Simple analytics tracking snippet
const analyticsSnippet = new webflow.InlineScript("analytics-snippet", {
  siteId: siteId,
  displayName: "AnalyticsSnippet",
  sourceCode: `(function() {
  window.dataLayer = window.dataLayer || [];
  function gtag() { dataLayer.push(arguments); }
  gtag('js', new Date());
  gtag('config', 'G-XXXXXXXXXX');
})();`,
  version: "1.0.0",
  canCopy: true, // Allow copying when site is duplicated
});

// Example 2: Cookie consent banner script
const cookieConsent = new webflow.InlineScript("cookie-consent", {
  siteId: siteId,
  displayName: "CookieConsent",
  sourceCode: `document.addEventListener('DOMContentLoaded', function() {
  if (!localStorage.getItem('cookieConsent')) {
    var banner = document.createElement('div');
    banner.id = 'cookie-banner';
    banner.innerHTML = '<p>We use cookies.</p><button id="accept-cookies">Accept</button>';
    document.body.appendChild(banner);
    document.getElementById('accept-cookies').addEventListener('click', function() {
      localStorage.setItem('cookieConsent', 'true');
      banner.remove();
    });
  }
});`,
  version: "1.2.0",
  canCopy: true,
});

// Example 3: Custom scroll-to-top button
const scrollToTop = new webflow.InlineScript("scroll-to-top", {
  siteId: siteId,
  displayName: "ScrollToTop",
  sourceCode: `document.addEventListener('DOMContentLoaded', function() {
  var btn = document.createElement('button');
  btn.textContent = '↑';
  btn.id = 'scroll-top-btn';
  btn.style.cssText = 'position:fixed;bottom:20px;right:20px;display:none;z-index:999;';
  document.body.appendChild(btn);
  window.addEventListener('scroll', function() {
    btn.style.display = window.scrollY > 300 ? 'block' : 'none';
  });
  btn.addEventListener('click', function() {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  });
});`,
  version: "2.0.0",
  canCopy: false, // Don't copy when duplicating site
});

// Export the script IDs and details for use in other resources
export const deployedSiteId = siteId;
export const analyticsSnippetId = analyticsSnippet.id;
export const cookieConsentId = cookieConsent.id;
export const scrollToTopId = scrollToTop.id;

// Export created and updated timestamps
export const analyticsSnippetCreatedOn = analyticsSnippet.createdOn;
export const analyticsSnippetLastUpdated = analyticsSnippet.lastUpdated;

// Print deployment success message
const message = pulumi.interpolate`Successfully registered 3 inline scripts to site ${siteId}`;
message.apply((m) => console.log(m));
