#!/usr/bin/env python3
"""
Post-process the generated Java SDK build.gradle for Maven Central publishing.

pulumi-java-gen generates a build.gradle with placeholder values that need to be
filled in for Maven Central compliance. This script patches the file with:
- Correct POM metadata (license, developer info, SCM URLs)
- gradle-nexus-publish-plugin for Maven Central Portal publishing
- Proper GPG signing configuration (3-param useInMemoryPgpKeys)
"""

import re
import sys
from pathlib import Path


class PatchError(Exception):
    """Raised when patching fails due to unexpected file format."""
    pass


def patch_build_gradle(filepath: str) -> None:
    """
    Patch the generated build.gradle file with Maven Central publishing configuration.

    Args:
        filepath: Path to the build.gradle file to patch

    Raises:
        FileNotFoundError: If the file doesn't exist
        PermissionError: If the file can't be read/written
        PatchError: If the file format is unexpected and patching fails
    """
    path = Path(filepath)

    if not path.exists():
        raise FileNotFoundError(f"build.gradle not found: {filepath}")

    try:
        content = path.read_text(encoding='utf-8')
    except PermissionError:
        raise PermissionError(f"Cannot read build.gradle (permission denied): {filepath}")

    original_content = content

    # Add nexus publish plugin if not present
    if 'io.github.gradle-nexus.publish-plugin' not in content:
        content = content.replace(
            'id("maven-publish")',
            'id("maven-publish")\n    id("io.github.gradle-nexus.publish-plugin") version "2.0.0"'
        )

    # Add signingKeyId variable if not present
    if 'def signingKeyId' not in content:
        content = content.replace(
            'def signingKey = System.getenv("SIGNING_KEY")',
            'def signingKeyId = System.getenv("SIGNING_KEY_ID")\ndef signingKey = System.getenv("SIGNING_KEY")'
        )

    # Add publishStagingURL with default if not present or missing default
    # Note: The staging URL is for Maven Central Portal's staging API (not legacy OSSRH)
    # See: https://central.sonatype.org/publish/publish-portal-api/
    staging_url = "https://ossrh-staging-api.central.sonatype.com/service/local/"

    if 'def publishStagingURL' not in content:
        content = content.replace(
            'def publishRepoUsername = System.getenv("PUBLISH_REPO_USERNAME")',
            f'def publishStagingURL = System.getenv("PUBLISH_STAGING_URL") ?: "{staging_url}"\ndef publishRepoUsername = System.getenv("PUBLISH_REPO_USERNAME")'
        )
    elif 'def publishStagingURL' in content:
        # publishStagingURL exists; check if it already has a default value
        after_def = content.split('def publishStagingURL', 1)[1]
        first_line = after_def.split('\n', 1)[0] if '\n' in after_def else after_def
        if '?:' not in first_line:
            # publishStagingURL exists but without default value
            content = content.replace(
                'def publishStagingURL = System.getenv("PUBLISH_STAGING_URL")',
                f'def publishStagingURL = System.getenv("PUBLISH_STAGING_URL") ?: "{staging_url}"'
            )

    # Update publishRepoURL default to Maven Central snapshots (only if no default exists)
    content = re.sub(
        r'def publishRepoURL = System\.getenv\("PUBLISH_REPO_URL"\)(?!\s*\?:)',
        'def publishRepoURL = System.getenv("PUBLISH_REPO_URL") ?: "https://central.sonatype.com/repository/maven-snapshots/"',
        content
    )

    # Fix artifactId
    content = content.replace(
        'artifactId = "webflow"',
        'artifactId = "pulumi-webflow"'
    )

    # Fix inceptionYear
    content = content.replace(
        'inceptionYear = ""',
        'inceptionYear = "2025"'
    )

    # Fix pom name - be specific to avoid matching other name fields
    content = re.sub(
        r'(pom \{[^}]*?)name = ""',
        r'\1name = "Pulumi Webflow Provider"',
        content,
        count=1
    )

    # Fix URL placeholders
    content = content.replace(
        'url = "https://example.com"',
        'url = "https://github.com/jdetmar/pulumi-webflow"'
    )
    content = content.replace(
        'connection = "https://example.com"',
        'connection = "scm:git:git://github.com/jdetmar/pulumi-webflow.git"'
    )
    content = content.replace(
        'developerConnection = "https://example.com"',
        'developerConnection = "scm:git:ssh://github.com:jdetmar/pulumi-webflow.git"'
    )

    # Fix license block - need to be careful not to match the pom name
    # First fix license name
    content = re.sub(
        r'(licenses \{[^}]*?license \{[^}]*?)name = ""',
        r'\1name = "Apache-2.0"',
        content
    )
    # Then fix license url
    content = re.sub(
        r'(licenses \{[^}]*?license \{[^}]*?)url = ""',
        r'\1url = "https://www.apache.org/licenses/LICENSE-2.0"',
        content
    )

    # Fix developer block
    content = re.sub(
        r'(developers \{[^}]*?developer \{[^}]*?)id = ""',
        r'\1id = "jdetmar"',
        content
    )
    content = re.sub(
        r'(developers \{[^}]*?developer \{[^}]*?)name = ""',
        r'\1name = "Justin Detmar"',
        content
    )
    content = re.sub(
        r'(developers \{[^}]*?developer \{[^}]*?)email = ""',
        r'\1email = "jdetmar@users.noreply.github.com"',
        content
    )

    # Fix signing to use 3 parameters
    content = content.replace(
        'useInMemoryPgpKeys(signingKey, signingPassword)',
        'useInMemoryPgpKeys(signingKeyId, signingKey, signingPassword)'
    )

    # Add nexusPublishing block if not present
    if 'nexusPublishing {' not in content:
        nexus_block = '''
if (publishRepoUsername) {
    nexusPublishing {
        repositories {
            sonatype {
                nexusUrl.set(uri(publishStagingURL))
                snapshotRepositoryUrl.set(uri(publishRepoURL))
                username = publishRepoUsername
                password = publishRepoPassword
            }
        }
    }
}
'''
        # Find the end of the publishing block and add nexusPublishing after it
        # Look for the closing brace of the publishing block
        publishing_match = re.search(r'(publishing \{.*?^\})\s*\n', content, re.MULTILINE | re.DOTALL)
        if publishing_match:
            insert_pos = publishing_match.end()
            content = content[:insert_pos] + nexus_block + content[insert_pos:]
        else:
            raise PatchError(
                "Failed to locate 'publishing { ... }' block in build.gradle; "
                "the file format may have changed and nexusPublishing could not be added. "
                f"File: {filepath}"
            )

    # Write the patched content
    try:
        path.write_text(content, encoding='utf-8')
    except PermissionError:
        raise PermissionError(f"Cannot write to build.gradle (permission denied): {filepath}")

    # Report what was done
    if content == original_content:
        print(f"No changes needed for {filepath} (already patched)")
    else:
        print(f"Successfully patched {filepath}")


def main() -> int:
    """Main entry point."""
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <build.gradle path>", file=sys.stderr)
        return 1

    try:
        patch_build_gradle(sys.argv[1])
        return 0
    except FileNotFoundError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1
    except PermissionError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1
    except PatchError as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1


if __name__ == '__main__':
    sys.exit(main())
