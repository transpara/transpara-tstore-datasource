# TStore Datasource

Visualize Transpara Platform process data — historian tags, asset model values, calculated KPIs — in Grafana dashboards alongside the rest of your observability stack.

This plugin queries `tstore-interface`, the Transpara Platform's REST API for time-series data, and renders it as native Grafana series so you can panel, alert, and Explore it like any other Grafana data source.

![Unit 1 Operations Overview](https://raw.githubusercontent.com/transpara/transpara-tstore-datasource/main/src/img/screenshot-dashboard-overview.png)

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

- Grafana **10.0.0** or newer, **self-hosted** (Grafana OSS or Enterprise).
  Grafana Cloud is **not supported** — Cloud only runs plugins from the Grafana catalog, and this plugin is distributed directly by Transpara.
- Filesystem and config access to the Grafana server (this plugin is installed by hand, not from the catalog).
- Network access from the Grafana server to:
  - A running `tstore-interface` deployment
  - The Keycloak realm fronting it
- A Keycloak **client** with the `client_credentials` grant enabled and appropriate `tstore-*` realm roles.

## Install

This plugin is **not published in the Grafana plugin catalog**. It is distributed directly by Transpara as an unsigned plugin and installed by copying it into your Grafana server's plugins directory ("side-loading"). This is a supported, documented Grafana installation path — see Grafana's [`allow_loading_unsigned_plugins`](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#allow_loading_unsigned_plugins) documentation.

Because the plugin is unsigned, Grafana will refuse to load it until you explicitly allow it by plugin ID. That is one config line, described in step 3.

### 1. Get the plugin

Download the latest release archive from the [releases page](https://github.com/transpara/transpara-tstore-datasource/releases), or build it from source (see [CONTRIBUTING.md](https://github.com/transpara/transpara-tstore-datasource/blob/main/CONTRIBUTING.md)).

### 2. Extract into the Grafana plugins directory

```bash
unzip transpara-tstore-datasource-<version>.zip -d /var/lib/grafana/plugins/
```

The resulting directory **must** be named `transpara-tstore-datasource` — it has to match the plugin `id` in `plugin.json`, or Grafana will not match it against the allow-list in the next step.

Make sure the files are readable by the Grafana service account and the backend binary is executable:

```bash
chown -R grafana:grafana /var/lib/grafana/plugins/transpara-tstore-datasource
chmod +x /var/lib/grafana/plugins/transpara-tstore-datasource/gpx_tstore-datasource*
```

Common plugin directories:

| Platform | Path |
|---|---|
| Linux package install | `/var/lib/grafana/plugins` |
| Docker / Kubernetes | `/var/lib/grafana/plugins` (mount a volume or bake into the image) |
| macOS (Homebrew) | `/opt/homebrew/var/lib/grafana/plugins` |
| Windows | `C:\Program Files\GrafanaLabs\grafana\data\plugins` |

### 3. Allow the unsigned plugin

Add the plugin ID to the unsigned allow-list in `grafana.ini` (or `custom.ini`):

```ini
[plugins]
allow_loading_unsigned_plugins = transpara-tstore-datasource
```

Or, equivalently, as an environment variable — the usual choice for Docker and Kubernetes:

```bash
GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=transpara-tstore-datasource
```

If you already allow other unsigned plugins, this is a comma-separated list:

```ini
allow_loading_unsigned_plugins = some-other-plugin,transpara-tstore-datasource
```

### 4. Restart Grafana

```bash
systemctl restart grafana-server
```

Then confirm under **Administration → Plugins → TStore Datasource** that the plugin is listed, and continue to [Configure](#configure).

### What you will see because the plugin is unsigned

None of these block anything once step 3 is done — they are expected:

| Where | What appears | What to do |
|---|---|---|
| Grafana server log, at startup | `Permitting unsigned plugin. This is not recommended` | Nothing. This is Grafana acknowledging your allow-list entry. |
| **Administration → Plugins** | An **Unsigned** badge next to TStore Datasource | Nothing. Cosmetic. |
| Data source config page | A warning banner noting the plugin is unsigned | Nothing. The data source works normally. |

If you **skipped or mistyped** step 3, the symptom is different: the plugin simply never appears in the data source list, and the server log shows

```
plugin registration failed ... error="plugin 'transpara-tstore-datasource' is unsigned"
```

Fix the plugin ID in `allow_loading_unsigned_plugins` (it must match the directory name exactly) and restart.

### Upgrading

Replace the directory contents with the new release and restart Grafana. The `allow_loading_unsigned_plugins` entry does not need to change.

### Air-gapped environments

Side-loading is the only install method, so air-gapped sites need no special handling: transfer the release archive across the boundary and follow the same four steps. No outbound network access to `grafana.com` is required at install time or at runtime.

## Configure

In Grafana, go to **Connections → Data sources → Add new data source** and select **TStore Datasource**.

| Field | Description | Example |
|---|---|---|
| **tstore-interface URL** | Base URL of the `tstore-interface` service (no trailing slash) | `https://your-server/tstore` |
| **Keycloak Token URL** | Full OIDC token endpoint | `https://your-server/tauth/realms/transpara/protocol/openid-connect/token` |
| **Client ID** | Keycloak service-account client ID | `textractor` |
| **Client Secret** | Keycloak client secret (stored encrypted by Grafana) | *(secret)* |

Click **Save & test**. A green "Datasource is working" toast confirms Keycloak auth and `tstore-interface` reachability end-to-end.


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

![Visual Query Editor](https://raw.githubusercontent.com/transpara/transpara-tstore-datasource/main/src/img/screenshot-query-editor.png)

### Raw mode

Toggle to **Raw** to send a JSON body directly to `POST /api/v1/read/trend-data`. Useful for queries that can't be expressed in the visual editor, or when you're iterating on a server-side feature. Switching from Visual → Raw seeds the JSON with the current visual selection so you can start from a known-good payload.

![Panel Detail — RCS Temperature, RCP Loop Flow, Pressurizer Level](https://raw.githubusercontent.com/transpara/transpara-tstore-datasource/main/src/img/screenshot-panel-detail.png)

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

- **Issues and feature requests:** [GitHub Issues](https://github.com/transpara/transpara-tstore-datasource/issues)
- **Documentation:** [docs.transpara.com](https://docs.transpara.com/)
- **Contact:** support@transpara.com

## License

MIT. See [LICENSE](https://github.com/transpara/transpara-tstore-datasource/blob/main/LICENSE).
