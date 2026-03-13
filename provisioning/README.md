# Provisioning

The `datasources/datasources.yml` file provisions the TStore datasource automatically on Grafana startup.

All connection settings are read from environment variables. Copy `.env.example` to `.env` and fill in your values before running `docker compose up`:

```bash
cp .env.example .env
```

| Variable | Description |
|---|---|
| `GF_DATASOURCE_URL` | tstore-interface base URL (no trailing slash) |
| `GF_DATASOURCE_TOKEN_URL` | Keycloak token endpoint URL |
| `GF_DATASOURCE_CLIENT_ID` | Keycloak client ID |
| `GF_DATASOURCE_CLIENT_SECRET` | Keycloak client secret |

For more information see [Provision dashboards and data sources](https://grafana.com/tutorials/provision-dashboards-and-data-sources/).
