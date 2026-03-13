# Testing the Plugin in Grafana

## Prerequisites

- Docker and Docker Compose
- Node.js >= 22
- Go >= 1.21 with [Mage](https://magefile.org/) installed
- Access to a running `tstore-interface` instance
- A Keycloak service account with `client_credentials` grant and access to tstore-interface

---

## Step 1: Build the plugin

```bash
npm install
npm run build
mage build:linuxARM64   # for Apple Silicon / ARM64 hosts
# or
mage build:linux        # for AMD64 hosts
```

Output lands in `dist/`.

---

## Step 2: Configure and run Grafana

Copy the example env file and fill in your connection details:

```bash
cp .env.example .env
```

Edit `.env`:

```
GF_DATASOURCE_URL=https://your-server/tstore
GF_DATASOURCE_TOKEN_URL=https://your-server/tauth/realms/transpara/protocol/openid-connect/token
GF_DATASOURCE_CLIENT_ID=textractor
GF_DATASOURCE_CLIENT_SECRET=your-client-secret
```

Then start Grafana:

```bash
docker compose up
```

This mounts `dist/` into Grafana's plugin directory, sets `allow_loading_unsigned_plugins`, and provisions the datasource automatically from the env vars above.

Grafana will be at **http://localhost:3000** (default login: `admin` / `admin`).

---

## Step 3: Verify the provisioned data source

1. Go to **Connections → Data Sources**
2. Open **TStore Datasource** — it should show the correct URLs from your `.env`
3. Click **Save & Test** — you should see the health check pass

If it fails, the error message will tell you whether it's a Keycloak auth failure or a tstore-interface connectivity issue.

---

## Step 4: Build a test dashboard

1. Go to **Dashboards → New → New Dashboard → Add visualization**
2. Select **TStore Datasource** as the data source
3. In the query editor:
   - **Visual mode**: Pick a dataset from the dropdown, select one or more lookups, choose an aggregation (e.g. `avg`) — leave Interval empty for auto
   - **Raw mode**: Toggle to Raw and paste a lookup array:
     ```json
     ["your-dataset|filter=value|avg|5m"]
     ```
4. Set the time range to a period where you have data
5. Click **Run query** — data should appear in the panel

---

## Step 5: Verify the lookup format

Lookups follow the format `dataset|filter` where:
- `filter` is comma-separated label pairs: `sensor_id=abc,unit=celsius`

In Visual mode, aggregation is controlled by the **Aggregation** and **Interval** fields in the query editor. In Raw mode, embed agg in the lookup: `dataset|filter|agg_type|agg_int`.

Supported aggregation types: `avg`, `min`, `max`, `sum`, `count`, `raw`, `median`, `twavg`

The visual mode dropdowns populate from the actual tstore-interface API so you can browse available datasets and lookups without knowing the format.

---

## Troubleshooting

| Symptom | Likely cause |
|---|---|
| "Save & Test" fails with auth error | Wrong client ID/secret or token URL |
| "Save & Test" fails with connection error | tstore-interface URL wrong or unreachable from Docker |
| Dataset dropdown is empty | Keycloak token valid but no datasets in tstore-interface |
| Lookup dropdown is empty after selecting dataset | No signals registered under that dataset |
| Panel shows "No data" | Time range has no data for that lookup |
| Plugin not showing in data source list | `dist/` not built, or wrong architecture binary (use `build:linuxARM64` on Apple Silicon) |
| "string does not match expected pattern" | Wrong binary architecture — rebuild with correct mage target |
