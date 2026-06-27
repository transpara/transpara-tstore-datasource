# Changelog

All notable changes to the TStore Datasource plugin are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] — 2026-06-11

Initial public release.

### Added

- Visual query editor with dataset and lookup dropdowns populated live from `tstore-interface`.
- Raw query mode that posts a JSON body directly to the `trend-data` endpoint for advanced use cases.
- Eight aggregation modes: `avg`, `min`, `max`, `sum`, `count`, `median`, `twavg`, and `raw`.
- Auto interval support — uses Grafana's `MaxDataPoints` for optimal panel-width resolution.
- Keycloak `client_credentials` authentication with in-memory token caching and automatic refresh on 401.
- Health check via `GET /api/v1/up` for end-to-end Save & test verification.
- Provisioning support for declarative deployment via environment variables.
- Signed plugin distribution.

### Compatibility

- Grafana **10.0.0** or newer.
- `tstore-interface` (Transpara Platform).
