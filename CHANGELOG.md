# Changelog

## 1.0.0

Initial release.

### Features

- Visual query editor with dataset and lookup dropdowns populated from tstore-interface
- Raw query mode for sending JSON directly to the trend-data endpoint
- Aggregation type selection: avg, min, max, sum, count, median, twavg, raw
- Auto interval: uses Grafana's panel width (`MaxDataPoints`) for optimal data resolution via `trend-data`
- Keycloak `client_credentials` authentication with token caching and automatic retry on 401
- Health check via `/api/v1/up`
- Provisioning support via environment variables
