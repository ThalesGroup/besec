# Security Policy

## Supported Versions

Only the latest version is supported. Security fixes will not be provided for
older versions.

## Reporting a Vulnerability

To report a vulnerability, please email us as described here:
https://cpl.thalesgroup.com/technical-support/how-to-report-a-security-vulnerability

## Security Update policy

Vulnerabilities will be communicated via GitHub Advisories and a description of
the issue will be included in the release notes.

## Security related configuration

Authentication is managed by an external identity provider. You may wish to run
this app within your private network / VPN, or behind a service such as Google
Identity Proxy.

Take care about white-listing an Identity Provider: this means that all users
who have an account with this identity provider will have access to your
instance. It is only suitable if the scope of the IdP and the target users of
the app align.

Once a user is authorized to use the app, there is no further access control:
all users can perform all actions. If you need to limit which users can modify
particular projects, PRs are welcome! Note that all admin tasks are via the
CLI, we will not accept contributions that add admin functionality to the web
app.
