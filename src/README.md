# TStore Datasource

Visualize Transpara Platform process data — historian tags, asset model values, calculated KPIs — in Grafana dashboards alongside the rest of your observability stack.

This plugin queries `tstore-interface`, the Transpara Platform's REST API for time-series data, and renders it as native Grafana series so you can panel, alert, and Explore it like any other Grafana data source.

<!-- TODO: insert screenshot of a panel with a real chart here once captured -->
<!-- ![Time-series panel](https://raw.githubusercontent.com/<org>/transpara-tstore-datasource/main/src/img/screenshots/panel.png) -->

## Features

- **Visual query editor** — pick a dataset, choose one or more lookups, set an aggregation and interval. No JSON required.
- **Raw query mode** — for advanced users, send a JSON body straight to the `trend-data` endpoint.
- **Eight aggregation modes** — `avg`, `min`, `max`, `sum`, `count`, `median`, `twavg` (time-weighted average), and `raw` for un-aggregated samples.
- **Auto interval** — pulls Grafana's `MaxDataPoints` from the panel width so the server returns exactly the resolution you can render.
- **Keycloak `client_credentials` auth** — service-account tokens are fetched, cached, and refreshed automatically; 401s trigger a transparent retry.
- **Backend plugin** — all auth and HTTP runs in the Grafana server, never in the browser. Secrets stay encrypted at rest.
- **Provisioning friendly** — every setting can be driven from environment variables for declarative deployments.
- **Health check** — `Save & test` round-trips Keycloak and `tstore-interface`, so misconfiguration surfaces immediately.

## Requirements

- Grafana **10.0.0** or newer (works on Grafana OSS, Enterprise, and Cloud).
- Network access from the Grafana server to:
  - A running `tstore-interface` deployment
  - The Keycloak realm fronting it
- A Keycloak **client** with the `client_credentials` grant enabled and appropriate `tstore-*` realm roles.

## Install

### From the Grafana plugin catalog

In Grafana, go to **Administration → Plugins**, search for **TStore Datasource**, and click **Install**. Restart is not required.

### Manual install (air-gapped)

1. Download the latest release archive from the [releases page](https://github.com/brettdbrewer/transpara-tstore-datasource/releases). <!-- TODO: update URL after org move -->
2. Extract into your Grafana plugins directory:
   ```bash
   unzip transpara-tstore-datasource-<version>.zip \
     -d /var/lib/grafana/plugins/
   ```
3. Restart the Grafana server.

The plugin is signed by Transpara; no `allow_loading_unsigned_plugins` override is required.

## Configure

In Grafana, go to **Connections → Data sources → Add new data source** and select **TStore Datasource**.

| Field | Description | Example |
|---|---|---|
| **tstore-interface URL** | Base URL of the `tstore-interface` service (no trailing slash) | `https://your-server/tstore` |
| **Keycloak Token URL** | Full OIDC token endpoint | `https://your-server/tauth/realms/transpara/protocol/openid-connect/token` |
| **Client ID** | Keycloak service-account client ID | `textractor` |
| **Client Secret** | Keycloak client secret (stored encrypted by Grafana) | *(secret)* |

Click **Save & test**. A green "Datasource is working" toast confirms Keycloak auth and `tstore-interface` reachability end-to-end.

<!-- TODO: insert screenshot of config page with green health check here -->

### Provisioning

All four settings can be driven from environment variables in a provisioning YAML:

```yaml
apiVersion: 1

datasources:
  - name: TStore Datasource
    type: transpara-tstore-datasource
    access: proxy
    jsonData:
      url: '${GF_DATASOURCE_URL}'
      tokenUrl: '${GF_DATASOURCE_TOKEN_URL}'
      clientId: '${GF_DATASOURCE_CLIENT_ID}'
    secureJsonData:
      clientSecret: '${GF_DATASOURCE_CLIENT_SECRET}'
```

See Grafana's [provisioning docs](https://grafana.com/docs/grafana/latest/administration/provisioning/) for the full lifecycle.

## Query

### Visual mode

The default editor. Three controls, no JSON:

- **Dataset** — select from the dropdown (populated by `tstore-interface`).
- **Lookups** — one or more time-series identifiers, filtered by the chosen dataset.
- **Aggregation** — `avg`, `min`, `max`, `sum`, `count`, `median`, `twavg`, or `raw`.
- **Interval** *(optional)* — e.g. `5m`, `1h`. Leave blank to let Grafana's auto-interval pick the best resolution for the panel width.

<!-- TODO: insert screenshot of visual query editor here -->

### Raw mode

Toggle to **Raw** to send a JSON body directly to `POST /api/v1/read/trend-data`. Useful for queries that can't be expressed in the visual editor, or when you're iterating on a server-side feature. Switching from Visual → Raw seeds the JSON with the current visual selection so you can start from a known-good payload.

<!-- TODO: insert screenshot of raw query editor here -->

### Time range and variables

The panel's time range and `$__interval` are passed through to `tstore-interface` automatically. Grafana dashboard variables can be referenced inside the dataset, lookup, and raw JSON fields and are interpolated server-side.

## How it works

```
┌─────────┐    HTTPS    ┌──────────────┐    HTTPS    ┌──────────────────┐
│ Grafana │ ──────────► │ TStore plugin│ ──────────► │ tstore-interface │
│ browser │             │   (Go, in    │             │   (Transpara     │
│         │             │   Grafana    │             │    Platform)     │
└─────────┘             │   server)    │             └──────────────────┘
                        └──────┬───────┘
                               │
                               │ client_credentials
                               ▼
                        ┌──────────────┐
                        │   Keycloak   │
                        └──────────────┘
```

The backend (Go) handles every outbound call. The browser never sees the Keycloak secret, never sees a Bearer token, and cannot reach `tstore-interface` directly. Tokens are cached in-process and re-fetched on expiry or 401.

## Support

- **Issues and feature requests:** [GitHub Issues](https://github.com/brettdbrewer/transpara-tstore-datasource/issues) <!-- TODO: update after org move -->
- **Documentation:** [docs.transpara.com](https://docs.transpara.com/)
- **Contact:** support@transpara.com

## License

MIT. See [LICENSE](https://github.com/brettdbrewer/transpara-tstore-datasource/blob/main/LICENSE). <!-- TODO: update URL after org move -->
