# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < 0.1.0 | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability, please do **NOT** open a public issue.

Instead, please send a report to:
- **Email:** [your-email@example.com]
- **GitHub Security Advisory:** [Create a private security advisory](https://github.com/Stoufiler/JellyWolProxy/security/advisories/new)

Please include:
- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Any suggested fixes (optional)

We will acknowledge your report within 48 hours and aim to release a fix within 7 days for critical issues.

## Security Measures

This project implements several security best practices:

- **Automated dependency updates** via Renovate
- **Vulnerability scanning** with govulncheck in CI/CD
- **SBOM generation** for each release
- **Minimal Docker images** based on `scratch`
- **Non-root user** in Docker containers
- **Go security best practices** (no CGO, trimpath, stripped binaries)

## Security Updates

Security updates are released as patch versions (e.g., v0.1.1 → v0.1.2) and announced via:
- GitHub Security Advisories
- Release notes
- Git tags

Subscribe to releases to stay informed.
