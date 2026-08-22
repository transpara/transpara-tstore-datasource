# Plugin Internals

This document covers non-obvious patterns in the codebase that a new developer needs to understand before making changes. Read this after [architecture.md](architecture.md).

---

## 401 Retry in `doRequest`

`pkg/plugin/datasource.go` — `doRequest()`

Every outbound HTTP request goes through `doRequest`, which implements a single automatic retry on HTTP 401:

```go
resp, err := ds.httpClient.Do(req)
if resp.StatusCode == http.StatusUnauthorized {
    ds.tokenCache.Reset()
    // rebuild request and retry once
}
```

Two subtleties worth knowing:

**Request body is consumed on first send.** `http.Request.Body` is an `io.Reader` — once read, it's gone. The retry works because bodies are created from `bytes.NewReader(...)`, which Go's HTTP client automatically sets as `GetBody`. On retry, `GetBody()` is called to get a fresh reader. If you ever add a request path that creates a body from a one-shot reader (e.g. an `os.File`), the retry will fail. Always use `bytes.NewReader` or `strings.NewReader` for request bodies in this codebase.

**Only one retry.** A second consecutive 401 is returned as an error. This prevents an infinite loop if the client credentials themselves are wrong.

---

## Token Cache

`pkg/plugin/auth.go` — `TokenCache`

```go
type TokenCache struct {
    mu        sync.Mutex
    token     string
    expiresAt time.Time
}
```

- Protected by a `sync.Mutex` — safe for concurrent `QueryData` calls
- Subtracts 30 seconds from Keycloak's `expires_in` to avoid using a token that expires mid-flight
- `Reset()` zeroes `expiresAt` so the next `GetToken()` fetches a fresh token — called by `doRequest` on 401
- One cache per `Datasource` instance — if two datasource configurations exist (e.g. dev and prod tstore), their tokens are fully isolated

---

## Interval Resolution

`pkg/plugin/query.go` — `runQuery()` and `formatDuration()`

Visual mode queries use a three-level interval fallback:

1. User-supplied `aggInt` (e.g. `"5m"`) — used verbatim if non-empty
2. Grafana's auto-computed `q.Interval` (`time.Duration`) — formatted with `formatDuration()`
3. `"5m"` hardcoded fallback if `q.Interval` is zero

`formatDuration()` converts a `time.Duration` to tstore's interval strings:
- Sub-minute durations → `"1m"` (minimum)
- Minutes → `"Nm"` (rounds down, e.g. 90s → `"1m"`)
- Hours (≥60 min) → `"Nh"` (e.g. 120m → `"2h"`)

The `aggType == "raw"` special case: if aggregation is set to `raw`, both `agg_type` and `agg_int` are **omitted from the request entirely** — not sent as `raw`, not sent as empty. tstore returns unaggregated data when these params are absent. The condition is:

```go
if qm.AggType != "" && qm.AggType != "raw" {
    params.Set("agg_type", qm.AggType)
    params.Set("agg_int", aggInt)
}
```

---

## Nullable Values in DataFrames

`pkg/plugin/query.go` — `toDataFrames()`

Data values use `*float64` (pointer to float64) rather than `float64`. This means:
- A missing or null data point is represented as `nil`, not `0`
- Grafana renders `nil` as a gap in the time series rather than a zero

tstore can return `v` as either a JSON number or a numeric string (`"42.5"`). Both are handled:

```go
switch val := p.V.(type) {
case float64:
    f = &val
case string:
    if parsed, err := strconv.ParseFloat(val, 64); err == nil {
        f = &parsed
    }
    // non-numeric string → nil (gap)
}
```

Non-numeric strings produce `nil`. This is intentional — tstore may return status strings for non-numeric sensors; treating them as gaps is safer than crashing.

---

## Frame Naming

`pkg/plugin/query.go` — `toDataFrames()`

A lookup string has the form `dataset|filter|...`. The Grafana panel legend shows the frame's `Name` field. The backend sets:

```go
frame.Name = parts[1]   // the filter segment, e.g. "sensor_id=motor-1,unit=rpm"
frame.RefID = key       // the full lookup string
```

This means the panel legend shows the filter portion, not the full `dataset|filter` string. If a lookup string has fewer than two `|`-separated segments, the frame name falls back to the full key.

---

## Lookups Endpoint: Paginated Response Unwrap

`pkg/plugin/datasource.go` — `CallResource()`

tstore's `/api/v1/lookups/` endpoint returns a paginated envelope:

```json
{ "results": ["...", "..."], "total_count": 42 }
```

The backend unwraps this before sending to the frontend — the frontend receives a plain `string[]`. This is done so the frontend doesn't need to know about tstore's pagination format. The limit is hard-coded at `1000` items per request; if a dataset has more than 1000 lookups, only the first 1000 are returned. This is currently a known limitation.

---

## Visual → Raw Mode Migration

`src/components/QueryEditor.tsx`

When a user switches from visual to raw mode, the current `lookups[]` array is serialized to `rawJson`:

```ts
onChange({ ...query, queryType: 'raw', rawJson: JSON.stringify(query.lookups) });
```

This is a one-way convenience — raw→visual does not attempt to parse `rawJson` back into `lookups[]` (the raw JSON may contain lookup strings with embedded agg params not expressible in the visual fields). Switching raw→visual gives the user an empty visual form.

---

## Frontend: Silent Dropdown Errors

`src/datasource.ts` — `getDatasets()` and `getLookups()`

Both methods catch all errors and return `[]`:

```ts
async getDatasets(): Promise<string[]> {
    try {
        return await getBackendSrv().fetch<string[]>(...).toPromise();
    } catch {
        return [];
    }
}
```

If tstore-interface is unreachable or returns an error, the dropdowns simply appear empty rather than showing an error banner in the query editor. This is intentional UX — a datasource connectivity error is surfaced by the health check, not by every dropdown. If you're debugging missing dropdown options, check the **Save & Test** result on the datasource config page first.

---

## ESM/CJS Jest Workaround

`jest.config.js` and `src/__mocks__/react-hookz-web.js`

`@grafana/ui` v13 pulls in `@react-hookz/web`, which is ESM-only (no CommonJS build). Jest's default jsdom environment requires CommonJS. The workaround has three parts:

1. **`jest.config.js`** — adds `@react-hookz/web` to `transformIgnorePatterns` so `@swc/jest` transforms it, and adds a `moduleNameMapper` entry redirecting it to the mock
2. **`src/__mocks__/react-hookz-web.js`** — a CJS stub that exports `useIsomorphicLayoutEffect: useEffect` (the only export used by the Grafana v13 packages)
3. **`jest-setup.js`** — polyfills `global.MessageChannel` from `worker_threads` because React 19 uses `MessageChannel` internally and jsdom doesn't expose it

If you upgrade `@grafana/*` in the future and tests start failing with `SyntaxError: Cannot use import statement outside a module` or `MessageChannel is not defined`, these three files are where to look.

---

## `.config/` is Off-limits

The `.config/` directory is generated and managed by `@grafana/create-plugin`. Modifying it directly will be overwritten by the next scaffold update (`npx @grafana/create-plugin@latest update`).

To extend behavior:
- **Webpack**: export a function from `webpack.config.ts` at root that merges with `.config/webpack/webpack.config.ts`
- **Jest**: extend `.config/jest.config` in root `jest.config.js` (already done — see that file)
- **ESLint**: extend in root `eslint.config.mjs`
- **TypeScript**: extend `.config/tsconfig.json` in root `tsconfig.json`

See [Grafana's extension guide](https://grafana.com/developers/plugin-tools/how-to-guides/extend-configurations.md) for details.

---

## go_plugin_build_manifest

`dist/go_plugin_build_manifest`

This file is generated by `mage buildAll` and lists SHA256 hashes of every Go source file that went into the build. Grafana's plugin validator uses it to verify the submitted binary matches the source code.

A past issue: when `.cache/` (npm's npx cache) was present during the build, the manifest generator swept up `.go` files from npm packages that ship Go implementations (specifically `node_modules/flatted/golang/pkg/flatted/flatted.go`). The fix is to delete `.cache/` before running `mage`. This directory is gitignored but can appear after running `npx` commands locally.

If the validator reports `go-manifest-issue`, run:
```bash
rm -rf .cache/
mage -v buildAll
```

Then repackage and resubmit.

---

## camelCase vs snake_case Field Names

The plugin bridges TypeScript (camelCase) and Python/tstore (snake_case). This is intentional and handled by the Go backend:

| TypeScript (query JSON) | Go struct field | tstore URL param |
|---|---|---|
| `aggType` | `AggType` | `agg_type` |
| `aggInt` | `AggInt` | `agg_int` |
| `queryType` | `QueryType` | — |
| `rawJson` | `RawJSON` | — |

The Go backend deserializes the TypeScript query JSON (camelCase) and serializes tstore HTTP params (snake_case). Do not change either side to match the other — the convention on each side is correct for that ecosystem.
