# transpara-tstore-datasource

Grafana data source plugin for [tstore-interface](../tstore-interface) (Transpara Platform).

## Build

```bash
npm install
npm run build        # frontend
mage -v build:backend   # Go backend (current platform)
mage -v buildAll        # frontend + Go backend (all platforms)
```

## Test

```bash
go test ./pkg/plugin/... -v -race   # Go backend tests
npm test -- --watchAll=false         # frontend tests
```

## Install (local / air-gapped)

1. Build the plugin: `mage -v build:backend`
2. Copy `dist/` to your Grafana plugin directory:
   ```bash
   cp -r dist/ /var/lib/grafana/plugins/transpara-tstore-datasource
   ```
3. Allow unsigned plugins in `grafana.ini`:
   ```ini
   [plugins]
   allow_loading_unsigned_plugins = transpara-tstore-datasource
   ```
4. Restart Grafana and add the data source under **Configuration → Data Sources → TStore Datasource**.

## Configuration

| Field | Description |
|---|---|
| URL | Base URL of tstore-interface (e.g. `https://tstore.internal:8080`) |
| Keycloak Token URL | Full token endpoint URL |
| Client ID | Keycloak service account client ID |
| Client Secret | Keycloak service account client secret (stored encrypted by Grafana) |

## Architecture

Frontend (React/TypeScript) + Go backend (Grafana plugin SDK).

- **Go backend** handles all auth (Keycloak client credentials flow) and HTTP calls to tstore-interface
- **Frontend** provides visual query editor (dataset/lookup dropdowns, aggregation controls) and raw JSON editor
- **Auth**: service account Keycloak token, cached and auto-refreshed on expiry or 401

## Query Modes

**Visual mode:** Select a dataset, pick lookups from a dropdown, choose aggregation function and interval.

**Raw mode:** Write the JSON body directly sent to `POST /api/v1/read/trend-data`. Toggle from visual→raw serializes your current selection as JSON.
