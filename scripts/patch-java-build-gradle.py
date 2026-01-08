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


def patch_build_gradle(filepath: str) -> None:
    with open(filepath, 'r') as f:
        content = f.read()

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
    if 'def publishStagingURL' not in content:
        content = content.replace(
            'def publishRepoUsername = System.getenv("PUBLISH_REPO_USERNAME")',
            'def publishStagingURL = System.getenv("PUBLISH_STAGING_URL") ?: "https://ossrh-staging-api.central.sonatype.com/service/local/"\ndef publishRepoUsername = System.getenv("PUBLISH_REPO_USERNAME")'
        )
    elif '?:' not in content.split('def publishStagingURL')[1].split('\n')[0]:
        # publishStagingURL exists but without default value
        content = content.replace(
            'def publishStagingURL = System.getenv("PUBLISH_STAGING_URL")',
            'def publishStagingURL = System.getenv("PUBLISH_STAGING_URL") ?: "https://ossrh-staging-api.central.sonatype.com/service/local/"'
        )

    # Update publishRepoURL default to Maven Central snapshots
    content = re.sub(
        r'def publishRepoURL = System\.getenv\("PUBLISH_REPO_URL"\)',
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

    with open(filepath, 'w') as f:
        f.write(content)

    print(f"Successfully patched {filepath}")


if __name__ == '__main__':
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} <build.gradle path>")
        sys.exit(1)
    patch_build_gradle(sys.argv[1])
