# Changelog

All notable changes to the TStore Datasource plugin are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **Distribution is side-loading only.** The plugin will not be published in the Grafana plugin catalog — catalog listing for a for-profit vendor requires Grafana's Commercial Plugin Subscription (US$60,000/year), which we have declined. Releases ship as unsigned archives installed by copying them into the Grafana plugins directory.
- Installation now requires `allow_loading_unsigned_plugins = transpara-tstore-datasource` in `grafana.ini` (or `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS`) on every install, not just air-gapped ones.
- Documentation updated throughout to describe side-loading as the only install path, including the expected unsigned-plugin warnings and how to resolve them.

### Removed

- Grafana Cloud support. Cloud only runs catalog plugins and ignores `allow_loading_unsigned_plugins`; this plugin requires self-hosted Grafana OSS or Enterprise.
- All catalog-submission and plugin-signing tooling: the `sign` npm script (`@grafana/sign-plugin`), the `Sign plugin` step and `GRAFANA_ACCESS_POLICY_TOKEN` env in `ci.yml`, and the `policy_token` input in `release.yml`. The corresponding Grafana access policy has been deleted.

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
- Distribution as an unsigned plugin, side-loaded into self-hosted Grafana.

### Compatibility

- Grafana **10.0.0** or newer.
- `tstore-interface` (Transpara Platform).
