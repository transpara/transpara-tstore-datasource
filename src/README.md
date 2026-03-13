# TStore Datasource

A Grafana backend datasource plugin for reading time-series data from the Transpara Platform's `tstore-interface` service.

## Requirements

- Grafana >= 10.0.0
- Access to a running `tstore-interface` instance
- A Keycloak client with `client_credentials` grant enabled

## Getting Started

### 1. Configure the datasource

In Grafana, go to **Connections → Data sources → Add new data source** and select **TStore Datasource**.

Fill in the following fields:

| Field | Description | Example |
|---|---|---|
| tstore-interface URL | Base URL of the tstore-interface service (no trailing slash) | `https://your-server/tstore` |
| Keycloak Token URL | Full token endpoint URL | `https://your-server/tauth/realms/transpara/protocol/openid-connect/token` |
| Client ID | Keycloak client ID | `textractor` |
| Client Secret | Keycloak client secret | *(keep secret)* |

Click **Save & test** to verify the connection.

### 2. Query data

In Explore or a dashboard panel, select the TStore datasource and use **Visual** mode:

- **Dataset** — select a dataset from the dropdown
- **Lookups** — select one or more time-series lookups (filtered by dataset)
- **Aggregation** — choose an aggregation function (`avg`, `min`, `max`, `sum`, `count`, `median`, `twavg`, or `raw`)
- **Interval** — aggregation interval (e.g. `5m`, `1h`); leave empty to use Grafana's auto interval

Use **Raw** mode to send a JSON body directly to the `trend-data` endpoint for advanced queries.

## Development

### Prerequisites

- Node.js >= 22
- Go >= 1.21
- [Mage](https://magefile.org/)

### Build

```bash
# Frontend
npm run build

# Backend (Linux AMD64)
mage build:linux

# Backend (Linux ARM64)
mage build:linuxARM64
```

### Run locally

```bash
cp .env.example .env
# Fill in your tstore-interface and Keycloak credentials in .env
docker compose up
```

Grafana will be available at http://localhost:3000.
