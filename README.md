# transpara-tstore-datasource

Grafana data source plugin for [tstore-interface](https://github.com/transpara/tstore-interface) (Transpara Platform).

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

## Install

The plugin is **not published in the Grafana plugin catalog**. Side-loading an unsigned build into a self-hosted Grafana is the only supported install method. Grafana Cloud cannot run it.

1. Build the plugin (or download a release archive):
   ```bash
   npm run build && mage -v buildAll
   ```
2. Copy `dist/` to your Grafana plugin directory, named exactly after the plugin ID:
   ```bash
   cp -r dist/ /var/lib/grafana/plugins/transpara-tstore-datasource
   chown -R grafana:grafana /var/lib/grafana/plugins/transpara-tstore-datasource
   ```
3. Allow the unsigned plugin in `grafana.ini`:
   ```ini
   [plugins]
   allow_loading_unsigned_plugins = transpara-tstore-datasource
   ```
   Or via environment variable (Docker/Kubernetes):
   ```bash
   GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=transpara-tstore-datasource
   ```
4. Restart Grafana and add the data source under **Connections → Data sources → TStore Datasource**.

Grafana will log `Permitting unsigned plugin. This is not recommended` at startup and show an **Unsigned** badge on the plugin page. Both are expected and harmless. If step 3 is missing or the ID is misspelled, the plugin will not load at all and the log will say `plugin 'transpara-tstore-datasource' is unsigned`.

Full end-user install guide, including per-platform plugin paths and upgrade steps: [src/README.md](src/README.md).

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
