# Development Guide

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | ≥ 1.27 | [go.dev/dl](https://go.dev/dl) or `brew install go` |
| Node.js | ≥ 22 | [nodejs.org](https://nodejs.org) or `brew install node` |
| mage | any | `go install github.com/magefile/mage@latest` |
| Docker + Compose | any | [docker.com](https://www.docker.com) |
| gh CLI | any | `brew install gh` (optional, for releases) |

Verify:
```bash
go version        # go1.27+
node --version    # v22+
mage --version
docker info
```

---

## Repository Layout

```
transpara-tstore-datasource/
├── src/                    # TypeScript/React frontend
│   ├── module.ts           # Webpack entry point
│   ├── datasource.ts       # Frontend DataSource class
│   ├── types.ts            # Shared TypeScript types
│   └── components/
│       ├── ConfigEditor.tsx
│       └── QueryEditor.tsx
├── pkg/                    # Go backend
│   ├── main.go
│   └── plugin/
│       ├── datasource.go   # CheckHealth, CallResource, doRequest
│       ├── query.go        # QueryData, toDataFrames
│       └── auth.go         # TokenCache, FetchToken
├── tests/                  # Playwright E2E tests
├── provisioning/           # Grafana provisioning YAML
├── dist/                   # Built plugin (committed for Docker dev)
├── .config/                # Grafana scaffolding — do not modify
├── Magefile.go             # Go build targets
├── go.mod / go.sum
├── package.json
├── jest.config.js          # Overrides scaffolding jest config
└── docker-compose.yaml
```

`.config/` is owned by Grafana's scaffolding toolchain and must not be modified directly. Extend webpack, jest, eslint, or prettier at the root level.

---

## Local Setup

```bash
git clone https://github.com/transpara/transpara-tstore-datasource
cd transpara-tstore-datasource
npm install
```

---

## Building

### Frontend

```bash
npm run build       # production build → dist/
npm run dev         # watch mode (rebuilds on file change)
npm run typecheck   # TypeScript type check only
npm run lint        # ESLint
npm run lint:fix    # ESLint + Prettier autofix
```

### Backend (Go)

```bash
mage -v buildAll          # all platforms (linux amd64/arm/arm64, windows amd64, darwin amd64/arm64)
mage -v build:backend     # current platform only (fastest for local iteration)
mage -v build:linux       # Linux AMD64
mage -v build:linuxARM64  # Linux ARM64 (required for Apple Silicon Docker)
mage -v build:darwinARM64 # macOS Apple Silicon
```

After building, ensure binaries are executable:
```bash
chmod 0755 dist/gpx_*
```

> **Apple Silicon note**: Docker runs Linux containers, so on an M1/M2/M3 Mac you must use `mage -v build:linuxARM64`, not `mage -v build:backend` (which produces a macOS binary that won't run inside the container).

---

## Running Tests

### Go Backend

```bash
go test ./pkg/plugin/... -v -race
```

Or via mage:
```bash
mage test
```

One test (`TestTokenCache_RefreshesOnExpiry`) sleeps for 2 seconds by design — it waits for a 1-second token to expire.

### Frontend Unit Tests

```bash
npm run test:ci       # single run (CI mode)
npm test              # watch mode
```

### End-to-End Tests (Playwright)

E2E tests require a running Grafana instance with the plugin loaded. Start the stack first:

```bash
cp .env.example .env   # fill in your credentials
npm run build
mage -v build:linuxARM64   # or build:linux on AMD64
npm run server             # docker compose up --build
```

Then in a second terminal:
```bash
npm run e2e
```

Grafana runs at `http://localhost:3000`. Default login: `admin` / `admin`.

---

## Local Docker Dev Environment

### First-time setup

```bash
cp .env.example .env
```

Edit `.env` with real values:
```
GRAFANA_VERSION=13.2.0
GRAFANA_IMAGE=grafana-enterprise
GF_DATASOURCE_URL=https://demo.transpara.io/tstore
GF_DATASOURCE_TOKEN_URL=https://keycloak.transpara.io/realms/transpara/protocol/openid-connect/token
GF_DATASOURCE_CLIENT_ID=your-client-id
GF_DATASOURCE_CLIENT_SECRET=your-client-secret
```

### Build and start

```bash
# Build frontend + backend for the container platform
npm run build
mage -v build:linuxARM64   # Apple Silicon
# or:
mage -v build:linux        # AMD64

# Start Grafana
docker compose up --build
```

Grafana: `http://localhost:3000` (admin/admin)
The datasource is provisioned automatically from your `.env` values — no manual setup needed.

### Iterating on frontend changes

```bash
npm run dev   # watch mode
```

The `dist/` directory is bind-mounted into the container. After webpack rebuilds, reload the Grafana page — no container restart needed for frontend-only changes.

### Iterating on backend changes

```bash
mage -v build:linuxARM64   # rebuild Go binary
docker compose restart grafana   # Grafana reloads the plugin binary on restart
```

### Viewing plugin logs

```bash
docker compose logs -f grafana | grep plugin.transpara
```

Log level is set to `debug` in the dev compose config (`GF_LOG_FILTERS: plugin.transpara-tstore-datasource:debug`).

### Wiping state

If Grafana's SQLite DB accumulates stale state (e.g. datasource config conflicts with provisioning):
```bash
docker compose down -v   # removes volumes, starts fresh next up
```

---

## Debugging (Delve)

The dev container runs Grafana with delve in headless mode on port `2345`. Two VS Code launch configs are provided in `.vscode/launch.json`:

- **Go: Launch** — builds and runs the plugin process standalone (outside Grafana, useful for unit-testing `main.go` flows)
- **Go: Remote Attach** — attaches to the delve instance inside the running Docker container

To use remote attach:
1. Start the stack: `npm run server`
2. In VS Code, select **Go: Remote Attach** and press F5

Breakpoints set in `pkg/plugin/` will be hit when Grafana calls the plugin.

---

## CI/CD Workflows

| Workflow | Trigger | What it does |
|---|---|---|
| `ci.yml` | Push / PR to `main` | typecheck → lint → unit tests → build → golangci-lint → mage buildAll → mage test → e2e matrix → validate plugin |
| `release.yml` | `v*` tag push | Full build + sign with `GRAFANA_ACCESS_POLICY_TOKEN` → GitHub release with attestation |
| `bundle-stats.yml` | Push / PR | Tracks frontend bundle size, comments on PRs |
| `is-compatible.yml` | PR | Grafana API compatibility check against latest Grafana |
| `cp-update.yml` | Monthly cron | Opens a PR with Grafana scaffolding updates |

### Creating a release

1. Update version in `src/plugin.json` and `package.json`
2. Update `CHANGELOG.md`
3. Commit and push to `main`
4. Tag the commit: `git tag v1.x.x && git push origin v1.x.x`
5. `release.yml` runs automatically, builds and signs the plugin, creates the GitHub release

The `GRAFANA_ACCESS_POLICY_TOKEN` secret must be set in the repository's GitHub Actions secrets. This token was created under the Transpara org on grafana.com with `plugin-submissions:write` scope.

---

## Updating the Grafana Scaffolding

The `.config/` directory is managed by `@grafana/create-plugin`. Grafana periodically releases updates. The `cp-update.yml` workflow opens monthly PRs to keep it in sync.

To manually apply an update:
```bash
npx @grafana/create-plugin@latest update
```

Review the diff carefully before merging — scaffolding changes can affect webpack, jest, TypeScript config, and Docker base images.

---

## Submitting to the Grafana Plugin Marketplace

See the [Grafana publishing docs](https://grafana.com/developers/plugin-tools/publish-a-plugin) for the full process. Summary:

1. Build the plugin: `npm run build && mage -v buildAll`
2. Package: `cp -r dist transpara-tstore-datasource && zip -qr transpara-tstore-datasource-<version>.zip transpara-tstore-datasource && rm -rf transpara-tstore-datasource`
3. Compute SHA1: `sha1sum transpara-tstore-datasource-<version>.zip`
4. Upload the zip to a GitHub release
5. Submit at `grafana.com/orgs/transpara/plugins/new` (must be admin of Transpara org)
6. After Grafana approves, sign: `GRAFANA_ACCESS_POLICY_TOKEN=<token> npm run sign`
7. Resubmit the signed zip
