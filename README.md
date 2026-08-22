# transpara-tstore-datasource

Grafana data source plugin for the [Transpara Platform](https://transpara.com). Connects Grafana dashboards to [tstore-interface](https://github.com/transpara/tstore-interface) (TimescaleDB REST API) using Keycloak service account authentication.

## Documentation

| Doc | Description |
|---|---|
| [Architecture](docs/architecture.md) | System context, component diagrams, data flow, auth state, lookup string format |
| [Development Guide](docs/development.md) | Prerequisites, build commands, local Docker dev, debugging, CI/CD, releasing |
| [Configuration Reference](docs/configuration.md) | All config fields, env vars, provisioning, common mistakes |
| [Plugin Internals](docs/internals.md) | Key patterns for new developers: 401 retry, token cache, interval resolution, nullable values, frame naming |

## Quick Start

```bash
npm install
npm run build                    # frontend
mage -v build:linuxARM64         # backend (Apple Silicon)
# or:
mage -v build:linux              # backend (AMD64)

cp .env.example .env             # fill in tstore/Keycloak credentials
docker compose up --build        # Grafana at http://localhost:3000
```

## Quick Test

```bash
go test ./pkg/plugin/... -v -race   # Go backend tests
npm run test:ci                      # frontend tests
```

## Plugin ID

`transpara-tstore-datasource`

Minimum Grafana version: `10.0.0`
