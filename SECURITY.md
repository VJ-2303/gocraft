# Security Policy

## Supported versions

Only the latest `main` and the most recent tagged release are supported with security fixes.

## Reporting a vulnerability

If you discover a security vulnerability, please **do not** open a public issue.

Email the maintainer privately or use GitHub's **Report a vulnerability** button on the repository's Security tab (advisory form). Include:

- A description of the vulnerability and its impact.
- Steps to reproduce, if applicable.
- Any suggested fix or mitigation.

You should receive an acknowledgement within 72 hours. Once the issue is confirmed, a fix will be prepared and a coordinated disclosure timeline agreed with you.

## Scope

This policy covers the `gocraft` tool itself. Codebases *generated* by `gocraft` are owned by their creators — vulnerabilities in generated dependency versions should be addressed by updating those dependencies in the generated project (see that project's `go.mod` and `make upgradeable`).
