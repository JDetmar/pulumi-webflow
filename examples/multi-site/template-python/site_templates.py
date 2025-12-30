"""Reusable site templates for Webflow multi-site management.

This module provides factory functions for creating standardized site
configurations that can be reused across your site fleet, ensuring
consistency and reducing configuration boilerplate.
"""

import pulumi_webflow as webflow


def create_campaign_site(name: str, display_name: str) -> webflow.Site:
    """Create a standardized campaign site with default configurations.

    Args:
        name: Unique identifier for the site (e.g., "q1-promo")
        display_name: Human-readable site name (e.g., "Q1 Promotion")

    Returns:
        Created Site resource
    """
    # Create site with campaign defaults
    site = webflow.Site(
        name,
        display_name=display_name,
        short_name=name.lower().replace(" ", "-"),
        time_zone="America/Los_Angeles",
    )

    # Standard robots.txt for campaigns (allow all)
    webflow.RobotsTxt(
        f"{name}-robots",
        site_id=site.id,
        content="User-agent: *\nAllow: /",
    )

    # Standard campaign redirects
    webflow.Redirect(
        f"{name}-home-redirect",
        site_id=site.id,
        source_path="/home",
        destination_path="/",
        status_code=301,
    )

    webflow.Redirect(
        f"{name}-signup-redirect",
        site_id=site.id,
        source_path="/join",
        destination_path="/signup",
        status_code=302,
    )

    return site


def create_product_site(name: str, display_name: str) -> webflow.Site:
    """Create a standardized product landing page with default configurations.

    Args:
        name: Unique identifier for the site (e.g., "product-alpha")
        display_name: Human-readable product name (e.g., "Product Alpha")

    Returns:
        Created Site resource
    """
    # Create site with product defaults
    site = webflow.Site(
        name,
        display_name=display_name,
        short_name=name.lower().replace(" ", "-"),
        time_zone="America/New_York",
    )

    # Product-specific robots.txt (allow all for indexing)
    webflow.RobotsTxt(
        f"{name}-robots",
        site_id=site.id,
        content="User-agent: *\nAllow: /",
    )

    # Standard product page redirects
    webflow.Redirect(
        f"{name}-pricing-redirect",
        site_id=site.id,
        source_path="/price",
        destination_path="/pricing",
        status_code=301,
    )

    webflow.Redirect(
        f"{name}-demo-redirect",
        site_id=site.id,
        source_path="/try",
        destination_path="/request-demo",
        status_code=302,
    )

    webflow.Redirect(
        f"{name}-docs-redirect",
        site_id=site.id,
        source_path="/help",
        destination_path="/documentation",
        status_code=301,
    )

    return site


def create_event_site(name: str, display_name: str) -> webflow.Site:
    """Create a standardized event microsite with default configurations.

    Args:
        name: Unique identifier for the site (e.g., "conference-2025")
        display_name: Event name (e.g., "Annual Conference 2025")

    Returns:
        Created Site resource
    """
    # Create site with event defaults
    site = webflow.Site(
        name,
        display_name=display_name,
        short_name=name.lower().replace(" ", "-"),
        time_zone="America/Chicago",
    )

    # Event-specific robots.txt
    webflow.RobotsTxt(
        f"{name}-robots",
        site_id=site.id,
        content="User-agent: *\nAllow: /\nDisallow: /admin/",
    )

    # Event registration redirects
    webflow.Redirect(
        f"{name}-register-redirect",
        site_id=site.id,
        source_path="/signup",
        destination_path="/register",
        status_code=301,
    )

    webflow.Redirect(
        f"{name}-agenda-redirect",
        site_id=site.id,
        source_path="/schedule",
        destination_path="/agenda",
        status_code=301,
    )

    webflow.Redirect(
        f"{name}-speakers-redirect",
        site_id=site.id,
        source_path="/presenters",
        destination_path="/speakers",
        status_code=301,
    )

    webflow.Redirect(
        f"{name}-tickets-redirect",
        site_id=site.id,
        source_path="/buy",
        destination_path="/tickets",
        status_code=301,
    )

    return site
