# Configuration Reference

## Datasource Fields

These fields are set in Grafana under **Connections → Data Sources → TStore Datasource → Settings**.

### jsonData (stored in Grafana DB, plaintext)

| UI Label | jsonData key | Required | Description |
|---|---|---|---|
| tstore-interface URL | `url` | Yes | Base URL of the tstore-interface service, no trailing slash. Example: `https://tstore.internal:8080` or `https://demo.transpara.io/tstore`. |
| Keycloak Token URL | `tokenUrl` | Yes | Full Keycloak token endpoint. Example: `https://keycloak.internal/realms/transpara/protocol/openid-connect/token` |
| Client ID | `clientId` | Yes | Keycloak service account client ID. Example: `tstore-grafana` or `textractor`. |

### secureJsonData (encrypted at rest by Grafana)

| UI Label | secureJsonData key | Required | Description |
|---|---|---|---|
| Client Secret | `clientSecret` | Yes | Keycloak service account client secret. Grafana encrypts this before storing and decrypts it before passing to the Go backend. Never stored in plaintext. |

### URL fallback

If `jsonData.url` is empty, the Go backend falls back to Grafana's built-in datasource `URL` field (`s.URL`). This allows the plugin to work with Grafana's standard proxy URL mechanism, but the preferred approach is to set `jsonData.url` explicitly.

---

## Environment Variables (provisioning)

Used exclusively in `provisioning/datasources/datasources.yml` with Grafana 13+'s `$__env{VAR}` interpolation syntax.

> **Note**: Grafana 13 changed the env var interpolation syntax from `${VAR}` to `$__env{VAR}`. The older syntax does not work in Grafana 13+ when `provisioning.gitConventions` is enabled (which is the default).

| Variable | Maps to | Example |
|---|---|---|
| `GF_DATASOURCE_URL` | `jsonData.url` | `https://demo.transpara.io/tstore` |
| `GF_DATASOURCE_TOKEN_URL` | `jsonData.tokenUrl` | `https://demo.transpara.io/tauth/realms/transpara/protocol/openid-connect/token` |
| `GF_DATASOURCE_CLIENT_ID` | `jsonData.clientId` | `textractor` |
| `GF_DATASOURCE_CLIENT_SECRET` | `secureJsonData.clientSecret` | (Keycloak client secret) |

Set these in `.env` for local Docker dev. `.env` is gitignored — never commit it.

---

## Provisioning File

`provisioning/datasources/datasources.yml`:

```yaml
apiVersion: 1

datasources:
  - name: TStore Datasource
    type: transpara-tstore-datasource
    access: proxy
    jsonData:
      url: '$__env{GF_DATASOURCE_URL}'
      tokenUrl: '$__env{GF_DATASOURCE_TOKEN_URL}'
      clientId: '$__env{GF_DATASOURCE_CLIENT_ID}'
    secureJsonData:
      clientSecret: '$__env{GF_DATASOURCE_CLIENT_SECRET}'
```

This file is mounted into `/etc/grafana/provisioning/datasources/` by Docker Compose. Grafana reads it on startup and creates (or updates) the datasource automatically. This means no manual configuration is needed in local dev — credentials come from `.env`.

In production Kubernetes deployments, inject the same four variables as environment variables into the Grafana pod (e.g. via a Kubernetes Secret mounted as env vars).

---

## Query Fields

These are set per panel in the query editor.

### Visual Mode

| Field | Query key | Description |
|---|---|---|
| Dataset | (dropdown selection) | The tstore dataset to query. Populated from `GET /api/v1/dataset`. |
| Lookups | `lookups[]` | One or more lookup strings identifying time series within the dataset. Multi-select dropdown. Populated from `GET /api/v1/lookups/?filter=<dataset>`. |
| Aggregation | `aggType` | Aggregation function: `avg`, `min`, `max`, `sum`, `count`, `median`, `twavg`, or `raw`. `raw` bypasses aggregation entirely and omits `agg_type` / `agg_int` from the tstore request. |
| Interval | `aggInt` | Aggregation interval string, e.g. `5m`, `1h`. Leave blank to use Grafana's auto-computed `$__interval` (recommended). Minimum 1 minute; falls back to `5m` if Grafana reports zero. |

### Raw Mode

| Field | Query key | Description |
|---|---|---|
| JSON body | `rawJson` | A JSON array of lookup strings sent verbatim as the POST body to `/api/v1/read/trend-data`. `agg_type` and `agg_int` are omitted from query params — embed them in the lookup strings if needed. |

Switching from visual to raw mode serializes your current `lookups[]` selection into the `rawJson` textarea as a convenience migration.

---

## Unsigned Plugin (local / air-gapped installs)

When installing the plugin outside of the Grafana marketplace, you must allow unsigned plugins in `grafana.ini`:

```ini
[plugins]
allow_loading_unsigned_plugins = transpara-tstore-datasource
```

Or via environment variable:
```
GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=transpara-tstore-datasource
```

The local Docker dev stack sets this automatically.

---

## Common Configuration Mistakes

**Wrong env var syntax in Grafana 13+**
Using `${VAR}` instead of `$__env{VAR}` in provisioning YAML. The value will be treated as a literal string rather than interpolated. Fix: use `$__env{VAR}`.

**Trailing slash in URL**
The Go backend strips trailing slashes, but it's cleaner to omit them in configuration. `https://tstore.internal:8080/` and `https://tstore.internal:8080` both work.

**Wrong Keycloak token URL**
The token URL must be the full path including `/protocol/openid-connect/token`, not just the realm base URL. The correct format is:
```
https://<keycloak-host>/realms/<realm>/protocol/openid-connect/token
```

**Stale provisioned datasource after credential change**
If you change credentials in `.env` and the datasource was previously provisioned, Grafana may cache the old config in its SQLite DB. Run `docker compose down -v` to wipe volumes and restart fresh.

**Client secret not updating after reset**
Clicking "Reset" in the ConfigEditor clears `secureJsonData.clientSecret` and sets `secureJsonFields.clientSecret: false`. You must type the new secret and save — just clicking Reset and saving without entering a new value will not update the stored secret.
