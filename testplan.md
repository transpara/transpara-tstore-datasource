# Testing the Plugin in Grafana

## Prerequisites

- Grafana 10+ running (Docker is easiest)
- `tstore-interface` running and reachable
- A Keycloak service account with access to tstore-interface

---

## Step 1: Build the plugin

```bash
cd /Users/brewskie/github/Grafana/transpara-tstore-datasource
npm install
npm run build
mage -v build:backend
```

Output lands in `dist/`.

---

## Step 2: Run Grafana with the plugin loaded

The repo includes a `docker-compose.yaml`. Set the required env var and start it:

```bash
export GF_DATASOURCE_CLIENT_SECRET=your-keycloak-client-secret
docker compose up
```

This mounts `dist/` into Grafana's plugin directory and sets `allow_loading_unsigned_plugins = transpara-tstore-datasource` automatically.

Grafana will be at **http://localhost:3000** (default login: `admin` / `admin`).

---

## Step 3: Add the data source

1. Go to **Connections → Data Sources → Add new data source**
2. Search for **TStore Datasource**
3. Fill in the fields:
   - **URL**: `http://your-tstore-host:8080`
   - **Keycloak Token URL**: `https://your-keycloak/realms/transpara/protocol/openid-connect/token`
   - **Client ID**: your Keycloak service account client ID
   - **Client Secret**: your client secret
4. Click **Save & Test** — you should see "Data source connected and auth verified."

If it fails, the error message will tell you whether it's a Keycloak auth failure or a tstore-interface connectivity issue.

---

## Step 4: Build a test dashboard

1. Go to **Dashboards → New → New Dashboard → Add visualization**
2. Select **TStore Datasource** as the data source
3. In the query editor:
   - **Visual mode**: Pick a dataset from the dropdown, select one or more lookups, choose an aggregation (e.g. `avg`) and interval (e.g. `5m`)
   - **Raw mode**: Toggle to Raw and paste a bare lookup array:
     ```json
     ["your-dataset|tag=value|avg|5m"]
     ```
4. Set the time range to a period where you have data
5. Click **Run query** — data should appear in the panel

---

## Step 5: Verify the lookup format

Lookups follow the format `dataset|filter|agg_type|agg_int` where:
- `filter` is comma-separated label pairs: `sensor_id=abc,unit=celsius`
- `agg_type` is one of: `avg`, `min`, `max`, `sum`, `count`, `raw`, `median`, `twavg`
- `agg_int` is a duration string like `5m`, `1h`, `30s`

The visual mode dropdowns populate from the actual tstore-interface API so you can browse available datasets and lookups without needing to know the format.

---

## Troubleshooting

| Symptom | Likely cause |
|---|---|
| "Save & Test" fails with auth error | Wrong client ID/secret or token URL |
| "Save & Test" fails with connection error | tstore-interface URL wrong or unreachable from Docker |
| Dataset dropdown is empty | Keycloak token valid but no datasets in tstore-interface |
| Lookup dropdown is empty after selecting dataset | No signals registered under that dataset |
| Panel shows "No data" | Time range has no data, or lookup string is invalid |
| Plugin not showing in data source list | `dist/` not built or Grafana not restarted after adding the plugin |
