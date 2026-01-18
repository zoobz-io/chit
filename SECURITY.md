# Security Policy

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          | Status |
| ------- | ------------------ | ------ |
| latest  | :white_check_mark: | Active development |
| < latest | :x: | Security fixes only for critical issues |

## Reporting a Vulnerability

We take the security of chit seriously. If you have discovered a security vulnerability in this project, please report it responsibly.

### How to Report

**Please DO NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **GitHub Security Advisories** (Preferred)
   - Go to the [Security tab](https://github.com/zoobzio/chit/security) of this repository
   - Click "Report a vulnerability"
   - Fill out the form with details about the vulnerability

2. **Email**
   - Send details to the repository maintainer through GitHub profile contact information
   - Use PGP encryption if possible for sensitive details

### What to Include

Please include the following information (as much as you can provide) to help us better understand the nature and scope of the possible issue:

- **Type of issue** (e.g., prompt injection, data leakage, stream hijacking, etc.)
- **Full paths of source file(s)** related to the manifestation of the issue
- **The location of the affected source code** (tag/branch/commit or direct URL)
- **Any special configuration required** to reproduce the issue
- **Step-by-step instructions** to reproduce the issue
- **Proof-of-concept or exploit code** (if possible)
- **Impact of the issue**, including how an attacker might exploit the issue
- **Your name and affiliation** (optional)

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Initial Assessment**: Within 7 days, we will provide an initial assessment of the report
- **Resolution Timeline**: We aim to resolve critical issues within 30 days
- **Disclosure**: We will coordinate with you on the disclosure timeline

### Preferred Languages

We prefer all communications to be in English.

## Security Best Practices

When using chit in your applications, we recommend:

1. **Keep Dependencies Updated**
   ```bash
   go get -u github.com/zoobzio/chit
   ```

2. **Protect LLM Credentials**
   - Never hardcode API keys in source code
   - Use environment variables or secret managers
   - Rotate keys regularly

3. **Validate LLM Outputs**
   - Treat LLM responses as untrusted input
   - Validate extracted data before using in critical paths
   - Use typed extraction (Extract[T]) for structured data

4. **Stream Security**
   - Implement authentication before providing Emitter access
   - Validate that clients are authorized to receive streamed data
   - Consider rate limiting on stream connections

5. **Context Management**
   - Be mindful of what data accumulates in Chat entries
   - Review what information flows between primitives
   - Use Fork primitives for content moderation routing

6. **Resource Management**
   - Use context timeouts for all operations
   - Implement rate limiting for LLM calls
   - Monitor token usage and costs

## Security Features

chit includes several built-in security features:

- **Type Safety**: Generic types prevent type confusion in extracted data
- **Context Support**: Built-in cancellation and timeout support
- **Error Isolation**: Errors are properly wrapped and traced
- **Observability**: Full signal emission for audit trails via capitan
- **Fork Primitives**: Binary routing for content validation and filtering

## LLM-Specific Considerations

When building LLM-powered chatbots:

- **Prompt Injection**: Validate and sanitize user inputs before including in prompts
- **Data Exfiltration**: Be cautious about what context is sent to external LLM providers
- **Hallucination**: Don't trust LLM outputs for security-critical decisions without validation
- **Cost Attacks**: Implement safeguards against requests designed to consume excessive tokens

## Automated Security Scanning

This project uses:

- **CodeQL**: GitHub's semantic code analysis for security vulnerabilities
- **Dependabot**: Automated dependency updates
- **golangci-lint**: Static analysis including security linters (gosec)
- **Codecov**: Coverage tracking to ensure security-critical code is tested

## Vulnerability Disclosure Policy

- Security vulnerabilities will be disclosed via GitHub Security Advisories
- We follow a 90-day disclosure timeline for non-critical issues
- Critical vulnerabilities may be disclosed sooner after patches are available
- We will credit reporters who follow responsible disclosure practices

## Credits

We thank the following individuals for responsibly disclosing security issues:

_This list is currently empty. Be the first to help improve our security!_

---

**Last Updated**: 2026-01-13
