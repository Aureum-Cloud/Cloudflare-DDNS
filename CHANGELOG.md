# Changelog

All notable changes to this project will be documented in this file.

## [1.0.2] - 2024-12-09

### Changed

- Fix SSL error by adding certificate authority for `scratch` image.

## [1.0.1] - 2024-11-09

### Added

- Add security policy (SECURITY.md).
- Add Helm chart for deploying Cloudflare DDNS to Kubernetes.

### Changed

- Reduced Docker image size by switching to `scratch` base image.

## [1.0.0] - 2024-11-03

### Added

- Initial commit with Cloudflare DDNS implementation.
